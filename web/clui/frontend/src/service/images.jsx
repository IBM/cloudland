/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import request from "../utils/request";
export function imagesListApi(paramsObj) {
  return request({
    url: "/api/images",
    method: "get",
    params: paramsObj ? paramsObj : {},
  });
}
export function createImgApi(objImg) {
  return request({
    url: "/api/images/new",
    method: "post",
    data: objImg,
  });
}
export function delImgInfor(imageid) {
  return request({
    url: `/api/images/${imageid}`,
    method: "delete",
    //data: registryid,
  });
}
