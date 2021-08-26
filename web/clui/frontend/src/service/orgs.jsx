/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import request from "../utils/request";

export function orgsListApi() {
  return request({
    url: "/api/orgs",
    method: "get",
  });
}
export function createOrgApi(objOrg) {
  return request({
    url: "/api/orgs/new",
    method: "post",
    data: objOrg,
  });
}
export function getOrgInforById(orgid) {
  return request({
    url: `/api/orgs/${orgid}`,
    method: "get",
  });
}
export function delOrgInfor(orgid) {
  return request({
    url: `/api/orgs/${orgid}`,
    method: "delete",
  });
}
export function editOrgInfor(orgid, obj) {
  return request({
    url: `/api/orgs/${orgid}`,
    method: "post",
    data: obj,
  });
}
