/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import React, { Component } from "react";
import { Card, Button, Popconfirm, message, Input } from "antd";
import { secrulesListApi, delSecruleInfor } from "../../service/secrules";
import DataTable from "../../components/DataTable/DataTable";

import "./secrules.css";
const { Search } = Input;

class Secrules extends Component {
  constructor(props) {
    super(props);
    console.log("Secrule.props:", this.props);
    this.state = {
      secrules: [],
      filteredList: [],
      sgID: "",
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
      key: "ID",
      width: 80,
      align: "center",
      // render: (href) => <a>{href}</a>,
    },
    {
      title: "SecurityGroup",
      dataIndex: "Secgroup",
      align: "center",
      // render: (href) => <a>{href}</a>,
    },
    {
      title: "RemoteIp",
      dataIndex: "RemoteIp",
      align: "center",
    },
    {
      title: "Direction",
      dataIndex: "Direction",
      align: "center",
    },
    {
      title: "Protocol",
      dataIndex: "Protocol",
      align: "center",
    },
    {
      title: "PortMin|Type",
      dataIndex: "PortMin",
      align: "center",
    },
    {
      title: "PortMax|Type",
      dataIndex: "PortMax",
      align: "center",
    },
    {
      title: "Action",
      align: "center",
      render: (txt, record, index) => {
        return (
          <div>
            <Button
              style={{
                marginTop: "10px",
              }}
              type="primary"
              size="small"
              onClick={() => {
                console.log("onClick:", record);
                console.log("onClick-Secgroup:", record.Secgroup);
                // this.setState({
                //   sgID: record.Secgroup,
                // });
                console.log("onClick-state.sgID:", this.state.sgID);
                this.props.history.push(
                  `/secgroups/${record.Secgroup}/secrules/new/` + record.ID
                  // `/secgroups/${this.state.sgID}/secrules/new/` + record.ID
                );
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
                console.log("onClick-delete-record.Secgroup:", record.Secgroup);

                delSecruleInfor(record.Secgroup, record.ID).then((res) => {
                  message.success(res.Msg);
                  this.loadData(this.state.current, this.state.pageSize);
                  console.log("用户~~", this.state);
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
              >
                Delete
              </Button>
            </Popconfirm>
          </div>
        );
      },
    },
  ];
  //组件初始化的时候执行
  componentDidMount() {
    const _this = this;
    console.log("componentWillMount:", this);
    secrulesListApi(this.props.match.params.id)
      .then((res) => {
        _this.setState({
          secrules: res.secrules,
          filteredList: res.secrules,
          isLoaded: true,
          total: res.total,
        });
        console.log(res);
      })
      .catch((error) => {
        _this.setState({
          isLoaded: false,
          error: error,
        });
      });
  }
  createSecrules = () => {
    console.log("createSecrules:", this.props);
    this.props.history.push(
      `/secgroups/${this.props.match.params.id}/secrules/new`
    );
  };
  listSecgroups = () => {
    this.props.history.push("/secgroups");
  };
  loadData = (page, pageSize) => {
    console.log("loadData~~", page, pageSize);
    const _this = this;
    const offset = (page - 1) * pageSize;
    const limit = pageSize;
    const sgID = this.props.match.params.id;
    secrulesListApi(sgID, { offset, limit })
      .then((res) => {
        console.log("loadData", res);
        _this.setState({
          secrules: res.secrules,
          filteredList: res.secrules,
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
    const sgID = this.props.match.params.id;

    // console.log("toSelectchange~limit:", sgID, offset, limit);
    secrulesListApi(sgID, { offset, limit })
      .then((res) => {
        console.log("loadData-toSelectchange", res);
        _this.setState({
          secrules: res.secrules,
          filteredList: res.secrules,
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
    console.log("onPaginationChange-pageSize", this.state.pageSize);
    this.loadData(e, this.state.pageSize);
  };
  onShowSizeChange = (current, pageSize) => {
    console.log("onShowSizeChange:", current, pageSize);
    //当几条一页的值改变后调用函数，current：改变显示条数时当前数据所在页；pageSize:改变后的一页显示条数
    this.toSelectchange(current, pageSize);
  };
  filter = (event) => {
    console.log("event-filter", event.target.value);
    this.getFilteredList(event.target.value);
  };
  getFilteredList = (word) => {
    console.log("getFilteredListr-keyword", word);
    var keyword = word.toLowerCase();
    if (keyword) {
      this.setState({
        filteredList: this.state.secrules.filter(
          (item) =>
            item.ID.toString().indexOf(keyword) > -1 ||
            item.Protocol.toLowerCase().indexOf(keyword) > -1 ||
            item.Direction.toLowerCase().indexOf(keyword) > -1 ||
            item.RemoteIp.indexOf(keyword) > -1
        ),
      });

      console.log("filteredList", this.state.filteredList);
    } else {
      this.setState({
        filteredList: this.state.secrules,
      });
    }
  };
  render() {
    return (
      <Card
        title={
          "Security Group Rules Manage Panel" +
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
                marginLeft: "50px",
                paddingLeft: "10px",
                paddingRight: "10px",
              }}
              type="primary"
              onClick={this.createSecrules}
            >
              Create
            </Button>
            <Button
              style={{
                float: "right",
                marginLeft: "10px",
                paddingLeft: "10px",
                paddingRight: "10px",
              }}
              type="primary"
              // size="small"
              onClick={this.listSecgroups}
            >
              Return
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

export default Secrules;
