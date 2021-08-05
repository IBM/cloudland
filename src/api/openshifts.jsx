import request from "../utils/request";

export function ocpListApi() {
  return request({
    url: "/api/openshifts",
    method: "get",
  });
}
