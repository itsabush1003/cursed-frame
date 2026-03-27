mergeInto(LibraryManager.library, {
  SplashDone: function() {
    window.dispatchReactUnityEvent("SplashDone");
  },
  QuestSceneReady: function() {
    window.dispatchReactUnityEvent("QuestSceneReady");
  },
  ProfileImageLoaded: function() {
    window.dispatchReactUnityEvent("ProfileImageLoaded");
  },
});