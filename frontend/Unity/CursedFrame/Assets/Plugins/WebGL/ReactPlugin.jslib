mergeInto(LibraryManager.library, {
  SplashDone: function() {
    window.dispatchReactUnityEvent("SplashDone");
  },
  QuestSceneReady: function() {
    window.dispatchReactUnityEvent("QuestSceneReady");
  },
  TargetMemberChanged: function() {
    window.dispatchReactUnityEvent("TargetMemberChanged");
  },
  TargetTeamChanged: function() {
    window.dispatchReactUnityEvent("TargetTeamChanged");
  },
  AttackAnimationEnd: function() {
    window.dispatchReactUnityEvent("AttackAnimationEnd");
  },
});