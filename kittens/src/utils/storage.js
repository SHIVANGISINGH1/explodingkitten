export const setItem = (key, data) => {
  console.log(data);
  return localStorage.setItem(key, JSON.stringify(data));
};

export const getItem = (key) => {
  return JSON?.parse(localStorage?.getItem(key));
};

export const removeItem = (key) => {
  localStorage?.removeItem(key);
};
