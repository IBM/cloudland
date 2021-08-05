import request from "../utils/request";

export function userListApi() {
  return request({
    url: "/api/users",
    method: "get",
  });
}
