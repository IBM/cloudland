import request from "../utils/request";

export function floatingipsListApi() {
  return request({
    url: "/api/floatingips",
    method: "get",
  });
}
