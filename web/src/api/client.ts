import axios, { type AxiosError } from "axios";

const client = axios.create({
  baseURL: "/api",
  withCredentials: true,
});

client.interceptors.response.use(
  res => res,
  (err: AxiosError) => {
    if (err.response?.status === 401) {
      window.location.href = "/login";
    }
    return Promise.reject(err);
  },
);

export default client;
