import request from "../utils/request";
export function imagesListApi() {
  return request({
    url: "/api/images",
    method: "get",
  });
}
