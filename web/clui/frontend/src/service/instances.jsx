/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import request from "../utils/request";

export function instListApi(paramsObj) {
  return request({
    url: "/api/instances/",
    method: "get",
    params: paramsObj ? paramsObj : {},
  });
}
export function createInstApi(objInst) {
  return request({
    url: "/api/instances/new",
    method: "post",
    data: objInst,
  });
}
export function getInstInforforAll() {
  return request({
    url: `/api/instances/new`,
    method: "get",
  });
}
export function getInstInforById(instid, paramsObj) {
  return request({
    url: `/api/instances/${instid}`,
    method: "get",
    params: paramsObj ? paramsObj : {},
  });
}
export function editInstInfor(instid, obj) {
  return request({
    url: `/api/instances/${instid}`,
    method: "post",
    data: obj,
  });
}
export function delInstInfor(instid) {
  return request({
    url: `/api/instances/${instid}`,
    method: "delete",
  });
}
