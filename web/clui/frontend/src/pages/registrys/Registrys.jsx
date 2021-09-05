/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import React, { Component } from "react";
import { Card, Button, Popconfirm, message, Input } from "antd";
import { regListApi, delRegInfor } from "../../service/registrys";
import DataTable from "../../components/DataTable/DataTable";

const { Search } = Input;
class Registrys extends Component {
  constructor(props) {
    super(props);
    console.log("props~~:", props);

    this.state = {
      registrys: [],
      filteredList: [],
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
                // this.gotoDeleteReg(record.ID);
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
    // const { regList } = this.props.reg;
    // const { handleFetchRegList } = this.props;
    // if (!regList || regList.length === 0) {
    //   handleFetchRegList();
    // }
    const _this = this;
    const limit = this.state.pageSize;
    regListApi(this.state.offset, limit)
      .then((res) => {
        console.log("regListApi-total:", res.total);
        _this.setState({
          filteredList: res.registrys,
          registrys: res.registrys,
          isLoaded: true,
          total: res.total,
        });
      })
      .catch((error) => {
        _this.setState({
          isLoaded: false,
          error: error,
        });
      });
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
          filteredList: res.registrys,
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
          filteredList: res.registrys,
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
  // gotoDeleteReg = (id) => {
  //   this.props.handleDeleteReg(id);
  // };
  filter = (event) => {
    console.log("event-filter", event.target.value);
    this.getFilteredList(event.target.value);
  };
  getFilteredList = (word) => {
    console.log("getFilteredListr-keyword", word);
    var keyword = word.toLowerCase();
    if (keyword) {
      this.setState({
        filteredList: this.state.registrys.filter(
          (item) =>
            item.ID.toString().indexOf(keyword) > -1 ||
            item.Label.toLowerCase().indexOf(keyword) > -1 ||
            item.OcpVersion.toLowerCase().indexOf(keyword) > -1 ||
            item.RegistryContent.toLowerCase().indexOf(keyword) > -1
        ),
      });

      console.log("filteredList", this.state.filteredList);
    } else {
      this.setState({
        filteredList: this.state.registrys,
      });
    }
  };

  render() {
    console.log("registry-props", this.props);

    return (
      <Card
        title={
          "Registry Manage Panel" +
          "(Total: " +
          this.state.filteredList.length +
          ")"
        }
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
                paddingLeft: "10px",
                paddingRight: "10px",
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
          dataSource={this.state.filteredList}
          bordered
          total={this.state.filteredList.length}
          pageSize={this.state.pageSize}
          scroll={{ y: 600 }}
          onPaginationChange={this.onPaginationChange}
          onShowSizeChange={this.onShowSizeChange}
          pageSizeOptions={this.state.pageSizeOptions}
          loading={!this.state.isLoaded}
        />
      </Card>
    );
  }
}

export default Registrys;
