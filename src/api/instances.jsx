import request from "../utils/request";
export function insListApi() {
  return request({
    url: "/api/instances",
    method: "get",
  });
}
export function createInstances(objInstance) {
  return request({
    url: "/api/instances/new",
    method: "post",
    data: objInstance,
  });
}
export function getInsInforById(instanceid) {
  return request({
    url: `/api/instances/${instanceid}`,
    method: "get",
  });
}
