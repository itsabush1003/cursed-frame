const LocalStorageRepository = {
  saveSecret: (reconnectKey: string) => {
    window.localStorage.setItem("key", reconnectKey);
  },
  getSecret: () => {
    const key = window.localStorage.getItem("key");
    return key;
  },
  removeSecret: () => {
    window.localStorage.removeItem("key");
  },
};

export default LocalStorageRepository;
