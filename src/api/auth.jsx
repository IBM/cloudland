import request from "../utils/request";

export function loginApi(user) {
  console.log("users:", user);
  return request({
    url: "/api/login",
    method: "post",
    data: user,
  });
}
