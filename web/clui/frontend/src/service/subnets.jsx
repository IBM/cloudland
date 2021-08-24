/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import request from "../utils/request";

export function subnetsListApi(paramsObj) {
  return request({
    url: "/api/subnets",
    method: "get",
    params: paramsObj ? paramsObj : {},
  });
}
export function createSubApi(objSub) {
  return request({
    url: "/api/subnets/new",
    method: "post",
    data: objSub,
  });
}
export function getSubInforById(subnetid) {
  return request({
    url: `/api/subnets/${subnetid}`,
    method: "get",
  });
}
export function delSubInfor(subnetid) {
  return request({
    url: `/api/subnets/${subnetid}`,
    method: "delete",
  });
}
export function editSubInfor(subnetid, objSub) {
  return request({
    url: `/api/subnets/${subnetid}`,
    method: "post",
    data: objSub,
  });
}
