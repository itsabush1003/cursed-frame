export const pathReplaceRegex = /\/static\/?$/;

export const waitExponentialBackoff = async (n: number) => {
  const delay = Math.pow(2, n);
  return new Promise((resolve) => setTimeout(() => resolve, delay));
}