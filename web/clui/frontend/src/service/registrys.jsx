/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import request from "../utils/request";
export function regListApi(paramsObj) {
  return request({
    url: "/api/registrys",
    method: "get",
    params: paramsObj ? paramsObj : {},
  });
}

export function createRegApi(objReg) {
  return request({
    url: "/api/registrys/new",
    method: "post",
    data: objReg,
  });
}
export function getRegInforById(registryid) {
  return request({
    url: `/api/registrys/${registryid}`,
    method: "get",
    //params: { registryid },
  });
}
export function delRegInfor(registryid) {
  return request({
    url: `/api/registrys/${registryid}`,
    method: "delete",
    //data: registryid,
  });
}
export function editRegInfor(registryid, obj) {
  return request({
    url: `/api/registrys/${registryid}`,
    method: "post",
    data: obj,
  });
}
