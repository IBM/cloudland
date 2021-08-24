/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import request from "../utils/request";

export function secrulesListApi(secgroupid, paramsObj) {
  return request({
    url: `/api/secgroups/${secgroupid}/secrules`,
    method: "get",
    params: paramsObj ? paramsObj : {},
  });
}
export function createSecruleApi(secgroupid, objSr) {
  return request({
    url: `/api/secgroups/${secgroupid}/secrules/new`,
    method: "post",
    data: objSr,
  });
}
export function getSecruleInforById(secgroupid, secruleid) {
  return request({
    url: `/api/secgroups/${secgroupid}/secrules/${secruleid}`,
    method: "get",
  });
}
export function delSecruleInfor(secgroupid, secruleid) {
  return request({
    url: `/api/secgroups/${secgroupid}/secrules/${secruleid}`,
    method: "delete",
  });
}
export function editSecruleInfor(secgroupid, secruleid, objSr) {
  return request({
    url: `/api/secgroups/${secgroupid}/secrules/${secruleid}`,
    method: "post",
    data: objSr,
  });
}
