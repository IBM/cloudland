/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import request from "../utils/request";

export function gwListApi(paramsObj) {
  return request({
    url: "/api/gateways",
    method: "get",
    params: paramsObj ? paramsObj : {},
  });
}
export function getGWInforById(gatewayid) {
  return request({
    url: `/api/gateways/${gatewayid}`,
    method: "get",
  });
}
export function createGWApi(objGW) {
  return request({
    url: "/api/gateways/new",
    method: "post",
    data: objGW,
  });
}
export function delGWInfor(gatewayid) {
  return request({
    url: `/api/gateways/${gatewayid}`,
    method: "delete",
  });
}
export function editGWInfor(gatewayid, obj) {
  return request({
    url: `/api/gateways/${gatewayid}`,
    method: "post",
    data: obj,
  });
}
