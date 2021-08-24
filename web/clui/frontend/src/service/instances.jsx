/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import request from "../utils/request";

export function insListApi(paramsObj) {
  return request({
    url: "/api/instances/",
    method: "get",
    params: paramsObj ? paramsObj : {},
  });
}
export function createInsApi(objInst) {
  return request({
    url: "/api/instances/new",
    method: "post",
    data: objInst,
  });
}
export function getInsInforById(instid) {
  return request({
    url: `/api/instances/${instid}`,
    method: "get",
  });
}
export function editInsInfor(instid, obj) {
  return request({
    url: `/api/instances/${instid}`,
    method: "post",
    data: obj,
  });
}
export function delInsInfor(instid) {
  return request({
    url: `/api/instances/${instid}`,
    method: "delete",
  });
}
