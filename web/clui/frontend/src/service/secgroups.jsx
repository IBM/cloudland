/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import request from "../utils/request";

export function secgroupsListApi(paramsObj) {
  return request({
    url: "/api/secgroups",
    method: "get",
    params: paramsObj ? paramsObj : {},
  });
}
export function createSecgroupApi(objSg) {
  return request({
    url: "/api/secgroups/new",
    method: "post",
    data: objSg,
  });
}
export function getSecgroupInforById(secgroupid) {
  return request({
    url: `/api/secgroups/${secgroupid}`,
    method: "get",
  });
}
export function delSecgroupInfor(secgroupid) {
  return request({
    url: `/api/secgroups/${secgroupid}`,
    method: "delete",
  });
}
export function editSecgroupInfor(secgroupid, objSg) {
  return request({
    url: `/api/secgroups/${secgroupid}`,
    method: "post",
    data: objSg,
  });
}
