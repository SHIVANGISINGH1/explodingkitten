import { handleError } from "./handleError";
import axios from "axios";

import { GET, POST, PUT } from "./httpConstants";
let hostUrl = process.env.REACT_APP_USER_API || "http://localhost:8000";
console.log(hostUrl);

export const request = (method, endPoint, reqBody = null) => {
  console.log(method);
  console.log("hostUrl", hostUrl + endPoint);
  const requestPromise = (httpMethod) => {
    return axios.request({
      url: hostUrl + endPoint,
      method: httpMethod,
      mode: "cors",
      headers: {
        "Content-Type": "application/json",
      },
      data: reqBody,
    });
  };

  switch (method) {
    case POST:
      return requestPromise(POST)
        .then((response) => response)
        .catch((error) => handleError(error));
    case PUT:
      return requestPromise(PUT)
        .then((response) => response)
        .catch((error) => handleError(error));
    case GET:
      return requestPromise(GET)
        .then((response) => response)
        .catch((error) => handleError(error));
    default:
      return "Wrong Call";
  }
};
