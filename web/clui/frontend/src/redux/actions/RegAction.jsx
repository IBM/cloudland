import {
  REG_LIST_FETCHED,
  LIST_FILTER,
  REG_LIST_IS_FETCHING,
} from "../../constants/actionTypes";
import { regListApi } from "../../service/registrys";
export const filterRegList = (keyword) => ({
  type: LIST_FILTER,
  keyword,
});
export const fetchRegList = () => {
  return (dispatch) => {
    dispatch(fetchingRegList(true));
    regListApi().then((res) => {
      console.log("regAction-res", res);
      // let resData = res.data;
      if (res) {
        dispatch(fetchingRegList(false));
        dispatch(fetchRegListSuccess(res.registrys, false));
        //   } else {
        //     //   dispatch(fetchingUserList(false));
        //     //   dispatch(fetchUserListFailed("获取用户列表失败"));
      }
    });
    //   .catch((e) => dispatch(fetchUserListFailed(e.message)));
  };
};
export const fetchRegListSuccess = (regList, loading) => ({
  type: REG_LIST_FETCHED,
  regList,
  loading,
});
export const fetchingRegList = (loading) => ({
  type: REG_LIST_IS_FETCHING,
  loading,
});
