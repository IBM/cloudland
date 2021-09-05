/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import { getAll } from "../../utils/auth";
const LoginReducer = (state = getAll(), action) => {
  switch (action.type) {
    case "LOGIN_INFO":
      return {
        loginInfo: state.loginInfo,
      };

    default:
      return state;
  }
};

export default LoginReducer;
