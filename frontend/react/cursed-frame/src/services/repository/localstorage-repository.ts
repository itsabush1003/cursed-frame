const LocalStorageRepository = {
  saveSecret: (reconnectKey: string, secret: string) => {
    window.localStorage.setItem("key", reconnectKey);
    window.localStorage.setItem("secret", secret);
  },
  getSecret: () => {
    const key = window.localStorage.getItem("key");
    const secret = window.localStorage.getItem("secret");
    return [key, secret];
  },
  removeSecret: () => {
    window.localStorage.removeItem("key");
    window.localStorage.removeItem("secret");
  },
};

export default LocalStorageRepository;
