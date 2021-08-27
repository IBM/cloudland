/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import React, { Component } from "react";
import moment from "moment";
import { Card, Table, Button, Popconfirm } from "antd";
import { orgsListApi, delOrgInfor } from "../../service/orgs";
import DataFilter from "../../components/Filter/DataFilter";

class Orgs extends Component {
  constructor(props) {
    super(props);
    this.state = {
      orgs: [],
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
    },
    {
      title: "Name",
      dataIndex: "name",
    },
    {
      title: "Created At",
      dataIndex: "CreatedAt",
    },
    {
      title: "Action",
      render: (txt, record, index) => {
        return (
          <div>
            <Button
              type="primary"
              size="small"
              onClick={() => {
                this.props.history.push("/orgs/new/" + record.ID);
              }}
            >
              Edit
            </Button>
            <Popconfirm
              title="Do you want to delete?"
              onCancel={() => {
                console.log("Cancel delete.");
              }}
              onConfirm={() => {
                console.log("Confirm delete.");
                delOrgInfor(record.ID).then((res) => {
                  message.success(res.Msg);
                  this.loadData(this.state.current, this.state.pageSize);
                });
              }}
            >
              <Button style={{ margin: "0 1rem" }} type="danger" size="small">
                Delete
              </Button>
            </Popconfirm>
          </div>
        );
      },
    },
  ];

  componentDidMount() {
    const _this = this;
    orgsListApi()
      .then((res) => {
        console.log("componentDidMount-orgsListApi:", res);
        _this.setState({
          orgs: res.orgs,
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

  loadData = (page, pageSize) => {
    const _this = this;
    const offset = (page - 1) * pageSize;
    const limit = pageSize;
    orgsListApi(offset, limit)
      .then((res) => {
        console.log("loadData", res);
        _this.setState({
          orgs: res.orgs,
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
    orgsListApi(offset, limit)
      .then((res) => {
        console.log("loadData", res);
        _this.setState({
          orgs: res.orgs,
          isLoaded: true,
          total: res.total,
          pageSize: limit,
          current: page,
        });
      })
      .catch((error) => {
        _this.setState({
          isLoaded: false,
          error: error,
        });
      });
  };

  createOrg = () => {
    this.props.history.push("/orgs/new");
  };

  render() {
    return (
      <Card
        title={
          "Organization Manage Panel" + "(Total: " + this.state.total + ")"
        }
        extra={
          <Button
            style={{
              float: "right",
              "padding-left": "10px",
              "padding-right": "10px",
            }}
            type="primary"
            onClick={this.createOrg}
          >
            Create
          </Button>
        }
      >
        <Table
          rowKey="ID"
          columns={this.columns}
          bordered
          dataSource={this.state.orgs}
          pagination={{
            //pagination
            total: this.state.total, //total count
            defaultPageSize: this.state.pageSize, //default pageSize
            showSizeChanger: true, //是否显示可以设置几条一页的选项
            onShowSizeChange: (current, pageSize) => {
              console.log("onShowSizeChange:", current, pageSize);
              //当几条一页的值改变后调用函数，current：改变显示条数时当前数据所在页；pageSize:改变后的一页显示条数
              this.toSelectchange(current, pageSize);
            },

            onChange: (current) => {
              this.loadData(current, this.state.pageSize);
            },
            showTotal: () => {
              return "Total " + this.state.total + " items";
            },
            pageSizeOptions: this.state.pageSizeOptions,
          }}
          loading={!this.state.isLoaded}
        ></Table>
      </Card>
    );
  }
}
export default Orgs;
