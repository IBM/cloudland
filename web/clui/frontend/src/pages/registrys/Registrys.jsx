/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import React, { Component } from "react";
import { Card, Button, Popconfirm, message, Input } from "antd";
import { regListApi, delRegInfor } from "../../service/registrys";
import DataTable from "../../components/DataTable/DataTable";
// import DataFilter from "../../components/Filter/DataFilter";
import { compose } from "redux";
import { withRouter } from "react-router";
import { connect } from "react-redux";
import { filterRegList, fetchRegList } from "../../redux/actions/RegAction";
const { Search } = Input;
class Registrys extends Component {
  constructor(props) {
    super(props);
    console.log("props~~:", props);

    this.state = {
      registrys: [],
      isLoaded: false,
      total: 0,
      pageSize: 10,
      offset: 0,
      pageSizeOptions: ["5", "10", "15", "20"],
      current: 1,
    };
  }
  columns = [
    {
      title: "ID",
      dataIndex: "ID",
      align: "center",
      key: "ID",
      className: "registry_id",
      width: 80,
    },
    {
      title: "Label",
      dataIndex: "Label",
      align: "center",
      width: 200,
      className: "registry_label",
    },
    {
      title: "Ocp Version",
      dataIndex: "OcpVersion",
      align: "center",
      className: "registry_ocpVersion",
      width: 80,
    },
    {
      title: "Registry Content",
      dataIndex: "RegistryContent",
      align: "center",
      className: "registry_Content",
      width: "45%",
      render: (text) => {
        if (text.length > 100) {
          return (
            <div
              className="registryContent"
              style={{
                overflow: "hidden",
                textOverflow: "ellipsis",
                display: "-webkit-box",
                WebkitBoxOrient: "vertical",
                WebkitLineClamp: "3",
                // maxWidth: 350,
              }}
            >
              {text}
            </div>
          );
        }
      },
    },
    {
      title: "Action",
      align: "center",
      // width: 160,
      render: (txt, record, index) => {
        return (
          <div>
            <Button
              style={{
                marginTop: "10px",
              }}
              type="primary"
              size="small"
              //onClick={() => console.log("onClick:", record)}
              onClick={() => {
                console.log("onClick:", record);
                this.props.history.push("/registrys/new/" + record.ID);
              }}
            >
              Edit
            </Button>
            <Popconfirm
              title="Are you sure to delete?"
              onCancel={() => {
                console.log("cancelled");
              }}
              onConfirm={() => {
                console.log("onClick-delete:", record);
                //this.props.history.push("/registrys/new/" + record.ID);
                delRegInfor(record.ID).then((res) => {
                  //const _this = this;
                  message.success(res.Msg);
                  this.loadData(this.state.current, this.state.pageSize);

                  console.log("用户~~", res);
                });
              }}
            >
              <Button
                style={{
                  margin: "5px",
                  marginRight: "0px",
                  marginTop: "10px",
                }}
                type="danger"
                size="small"
                onClick={() => {
                  console.log("record", record);
                  console.log("text", txt);
                  console.log("index", index);
                }}
              >
                Delete
              </Button>
            </Popconfirm>
          </div>
        );
      },
    },
  ];
  componentDidMount() {
    console.log("组件加载完成===================================");
    const { regList } = this.props.reg;
    const { handleFetchRegList } = this.props;
    if (!regList || regList.length === 0) {
      handleFetchRegList();
    }
  }

  createRegistrys = () => {
    this.props.history.push("/registrys/new");
  };
  loadData = (page, pageSize) => {
    console.log("loadData~~", page, pageSize);
    const _this = this;
    const offset = (page - 1) * pageSize;
    const limit = pageSize;
    regListApi(offset, limit)
      .then((res) => {
        console.log("loadData", res);

        _this.setState({
          registrys: res.registrys,
          isLoaded: true,
          total: res.total,
          pageSize: limit,
          current: page,
        });
        console.log("loadData-page-", page, _this.state);
      })
      .catch((error) => {
        _this.setState({
          isLoaded: false,
          error: error,
        });
      });
  };

  toSelectchange = (page, num) => {
    console.log("toSelectchange", page, num);
    const _this = this;
    const offset = (page - 1) * num;
    const limit = num;
    console.log("toSelectchange~limit:", offset, limit);
    regListApi(offset, limit)
      .then((res) => {
        console.log("loadData", res);
        _this.setState({
          registrys: res.registrys,
          isLoaded: true,
          total: res.total,
          pageSize: limit,
        });
      })
      .catch((error) => {
        _this.setState({
          isLoaded: false,
          error: error,
        });
      });
  };
  onPaginationChange = (e) => {
    console.log("onPaginationChange", e);
    this.loadData(e, this.state.pageSize);
  };
  onShowSizeChange = (current, pageSize) => {
    console.log("onShowSizeChange:", current, pageSize);
    //当几条一页的值改变后调用函数，current：改变显示条数时当前数据所在页；pageSize:改变后的一页显示条数
    this.toSelectchange(current, pageSize);
  };

  filter = (event) => {
    this.props.handleFilterRegList(event.target.value);
  };

  render() {
    console.log("registry-props", this.props);
    const { filteredList, isLoading } = this.props.reg;

    return (
      <Card
        title={"Registry Manage Panel" + "(Total: " + filteredList.length + ")"}
        extra={
          <div>
            <Search
              placeholder="Search..."
              onChange={this.filter}
              enterButton
            />
            <Button
              style={{
                float: "right",
                "padding-left": "10px",
                "padding-right": "10px",
              }}
              type="primary"
              onClick={this.createRegistrys}
            >
              Create
            </Button>
          </div>
        }
      >
        <DataTable
          rowKey="ID"
          columns={this.columns}
          dataSource={filteredList}
          bordered
          total={filteredList.length}
          pageSize={this.state.pageSize}
          scroll={{ y: 600 }}
          onPaginationChange={this.onPaginationChange}
          onShowSizeChange={this.onShowSizeChange}
          pageSizeOptions={this.state.pageSizeOptions}
          // loading={!this.state.isLoaded}
          loading={isLoading}
        />
      </Card>
    );
  }
}
const mapStateToProps = ({ reg }) => {
  console.log("mapStateToProps-state", reg);
  return {
    reg,
  };
};

const mapDispatchToProps = (dispatch) => {
  return {
    handleFetchRegList: () => dispatch(fetchRegList()),
    handleFilterRegList: (keyword) => dispatch(filterRegList(keyword)),
  };
};
export default compose(
  withRouter,
  connect(mapStateToProps, mapDispatchToProps)
)(Registrys);
