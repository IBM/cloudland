/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/

import request from "../utils/request";

//login api
export function loginApi(user) {
  return request({
    url: "/api/login",
    method: "post",
    data: user,
  });
}
