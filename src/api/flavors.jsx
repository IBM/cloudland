import request from "../utils/request";
export function flavorsListApi() {
  return request({
    url: "/api/flavors",
    method: "get",
  });
}
