// Drives internal/shared/layout.Dashboard's persistent player bar
// (#player-bar). The bar itself never leaves the DOM across boosted
// navigation (see PlayerBar's doc comment), so this module creates the
// YouTube IFrame player once and reuses it via loadVideoById for every
// later track instead of tearing the iframe down — playback survives
// switching dashboard pages.
//
// #now-playing-info (internal/features/playlist/views/now_playing.templ)
// is the source of truth for "what should be playing": its hx-trigger="load"
// fires once on first page load, and advance() below re-POSTs the same
// endpoint (with the just-finished track's id) every time playback needs to
// move on. This module only reacts to that element's data-youtube-id
// changing — it never decides "next" itself, so swapping the temporary
// created_at-order NextTrack for a real queue later needs no JS changes.
(function () {
  var RETRY_DELAY_MS = 5000;
  var AUTOPLAY_BLOCKED_TIMEOUT_MS = 3000;
  var PROGRESS_POLL_MS = 500;

  var bar = document.getElementById("player-bar");
  var nowPlayingInfo = document.getElementById("now-playing-info");
  var mount = document.getElementById("yt-player-mount");
  var progressFill = document.querySelector("#player-progress .progress-bar__fill");
  var timeCurrent = document.getElementById("player-time-current");
  var timeDuration = document.getElementById("player-time-duration");

  if (!bar || !nowPlayingInfo || !mount) return;

  function formatDuration(seconds) {
    seconds = Math.max(0, Math.floor(seconds || 0));
    var m = Math.floor(seconds / 60);
    var s = seconds % 60;
    return m + ":" + (s < 10 ? "0" : "") + s;
  }

  var player = null;
  var loadedYoutubeId = null;
  var isPlaying = false;
  var isMuted = false;
  var progressInterval = null;
  var autoplayBlockedTimer = null;

  var apiLoadPromise = null;
  function loadYouTubeIframeApi() {
    if (apiLoadPromise) return apiLoadPromise;

    apiLoadPromise = new Promise(function (resolve) {
      if (window.YT && window.YT.Player) {
        resolve();
        return;
      }
      var previousReady = window.onYouTubeIframeAPIReady;
      window.onYouTubeIframeAPIReady = function () {
        if (previousReady) previousReady();
        resolve();
      };
      var script = document.createElement("script");
      script.src = "https://www.youtube.com/iframe_api";
      document.head.appendChild(script);
    });

    return apiLoadPromise;
  }

  function setState(state) {
    if (state) {
      bar.setAttribute("data-state", state);
    } else {
      bar.removeAttribute("data-state");
    }
  }

  function clearAutoplayBlockedTimer() {
    if (autoplayBlockedTimer) {
      clearTimeout(autoplayBlockedTimer);
      autoplayBlockedTimer = null;
    }
  }

  function armAutoplayBlockedTimer() {
    clearAutoplayBlockedTimer();
    autoplayBlockedTimer = setTimeout(function () {
      if (!isPlaying) setState("blocked");
    }, AUTOPLAY_BLOCKED_TIMEOUT_MS);
  }

  function startProgressPolling() {
    stopProgressPolling();
    progressInterval = setInterval(function () {
      if (!player) return;
      var duration = player.getDuration();
      var current = player.getCurrentTime();
      if (timeCurrent) timeCurrent.textContent = formatDuration(current);
      if (timeDuration) timeDuration.textContent = formatDuration(duration);
      if (!duration || !progressFill) return;
      var pct = Math.max(0, Math.min(100, (current / duration) * 100));
      progressFill.style.width = pct + "%";
    }, PROGRESS_POLL_MS);
  }

  function stopProgressPolling() {
    if (progressInterval) {
      clearInterval(progressInterval);
      progressInterval = null;
    }
  }

  // advance re-requests #now-playing-info with the track that just finished
  // (or errored) so the server can resolve the next one — see the module
  // doc comment. A failed request retries after RETRY_DELAY_MS instead of
  // leaving the player silently stalled forever.
  function advance() {
    if (!window.htmx) {
      setTimeout(advance, RETRY_DELAY_MS);
      return;
    }

    // select/pushUrl mirror the declarative attributes on #now-playing-info's
    // own hx-post (see internal/shared/layout.PlayerBar) — without them this
    // call would still inherit hx-select="#dashboard-content"/hx-push-url="true"
    // from the boosted wrapper it sits inside, same trap that attributes
    // override for the "load" trigger.
    window.htmx.ajax("POST", "/playlist/now-playing/next", {
      target: "#now-playing-info",
      select: "#now-playing-info",
      swap: "outerHTML",
      pushUrl: false,
      values: { currentYoutubeId: loadedYoutubeId || "" },
    }).catch(function () {
      setTimeout(advance, RETRY_DELAY_MS);
    });
  }

  function loadTrack(youtubeId) {
    loadedYoutubeId = youtubeId;
    clearAutoplayBlockedTimer();
    setState(null);

    if (!player) {
      loadYouTubeIframeApi().then(function () {
        if (!window.YT || loadedYoutubeId !== youtubeId) return;

        player = new window.YT.Player(mount, {
          videoId: youtubeId,
          playerVars: { autoplay: 1, controls: 0, disablekb: 1, modestbranding: 1 },
          events: {
            onReady: function () {
              armAutoplayBlockedTimer();
            },
            onStateChange: function (event) {
              var playing = event.data === window.YT.PlayerState.PLAYING;
              isPlaying = playing;
              if (playing) {
                clearAutoplayBlockedTimer();
                setState("playing");
                startProgressPolling();
              } else {
                setState(null);
                stopProgressPolling();
              }
              if (event.data === window.YT.PlayerState.ENDED) {
                advance();
              }
            },
            onError: function () {
              advance();
            },
          },
        });
      });
      return;
    }

    player.loadVideoById(youtubeId);
    armAutoplayBlockedTimer();
  }

  // Not filtered by event.target: with hx-swap="outerHTML" htmx fires
  // htmx:afterSwap on the swapped element's *parent* (#player-bar here),
  // not on #now-playing-info itself, since the original target node was
  // removed from the DOM by the swap — see htmx's hx-swap docs. Re-reading
  // #now-playing-info fresh from the DOM on every swap (cheap, and a no-op
  // when its data-youtube-id hasn't changed) sidesteps that quirk instead
  // of trying to match the event's target.
  document.body.addEventListener("htmx:afterSwap", function () {
    var current = document.getElementById("now-playing-info");
    if (!current) return;

    var youtubeId = current.getAttribute("data-youtube-id");

    if (!youtubeId) {
      loadedYoutubeId = null;
      if (player) player.stopVideo();
      setState(null);
      stopProgressPolling();
      return;
    }

    if (youtubeId !== loadedYoutubeId) loadTrack(youtubeId);
  });

  document.body.addEventListener("htmx:responseError", function (event) {
    var xhr = event.detail && event.detail.xhr;
    if (xhr && xhr.responseURL && xhr.responseURL.indexOf("/playlist/now-playing/next") !== -1) {
      setTimeout(advance, RETRY_DELAY_MS);
    }
  });

  window.__togglePlayerPlayback = function () {
    if (!player) return;
    if (isPlaying) {
      player.pauseVideo();
    } else {
      player.playVideo();
    }
  };

  window.__togglePlayerMute = function () {
    if (!player) return;
    if (isMuted) {
      player.unMute();
    } else {
      player.mute();
    }
    isMuted = !isMuted;
    bar.classList.toggle("player-bar--muted", isMuted);
  };
})();
