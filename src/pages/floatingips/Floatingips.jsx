/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import React, { Component } from "react";
import { Card, Button, Popconfirm, message } from "antd";
import {
  floatingipsListApi,
  delFloatingipInfor,
} from "../../service/floatingips";
import DataTable from "../../components/DataTable/DataTable";
import DataFilter from "../../components/Filter/DataFilter";

class Floatingips extends Component {
  constructor(props) {
    super(props);
    this.state = {
      floatingips: [],
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
      title: "FloatingIP",
      dataIndex: "FipAddress",
      align: "center",
    },
    {
      title: "InternalIP",
      dataIndex: "IntAddress",
      align: "center",
    },
    {
      title: "Instance",
      dataIndex: "Instance.Hostname",
      align: "center",
    },
    {
      title: "Zone",
      dataIndex: "Instance.Zone.Name",
      align: "center",
    },
    {
      title: "Action",
      align: "center",
      render: (txt, record, index) => {
        return (
          <div>
            <Popconfirm
              title="Are you sure to delete?"
              onCancel={() => {
                console.log("cancell to delete");
              }}
              onConfirm={() => {
                console.log("onClick-delete-fl:", record);
                //this.props.history.push("/registrys/new/" + record.ID);
                delFloatingipInfor(record.ID).then((res) => {
                  //const _this = this;
                  message.success(res.Msg);
                  this.loadData(this.state.current, this.state.pageSize);

                  console.log("用户~~-fl", res);
                  console.log("用户~~state", this.state);
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
  createFloatingips = () => {
    this.props.history.push("/floatingips/new");
  };
  //组件初始化的时候执行
  componentDidMount() {
    const _this = this;
    console.log("componentDidMount:", this);
    floatingipsListApi()
      .then((res) => {
        _this.setState({
          floatingips: res.floatingips,
          isLoaded: true,
          total: res.total,
        });
        console.log("floatingipsListApi", res);
      })
      .catch((error) => {
        _this.setState({
          isLoaded: false,
          error: error,
        });
      });
  }
  loadData = (page, pageSize) => {
    console.log("image-loadData~~", page, pageSize);
    const _this = this;
    const offset = (page - 1) * pageSize;
    const limit = pageSize;
    floatingipsListApi(offset, limit)
      .then((res) => {
        console.log("loadData", res);

        _this.setState({
          floatingips: res.floatingips,
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
    console.log("flavor-toSelectchange~limit:", offset, limit);
    floatingipsListApi(offset, limit)
      .then((res) => {
        console.log("loadData", res);
        _this.setState({
          floatingips: res.floatingips,
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
  onPaginationChange = (e) => {
    console.log("onPaginationChange", e);
    this.loadData(e, this.state.pageSize);
  };
  onShowSizeChange = (current, pageSize) => {
    console.log("onShowSizeChange:", current, pageSize);
    //当几条一页的值改变后调用函数，current：改变显示条数时当前数据所在页；pageSize:改变后的一页显示条数
    this.toSelectchange(current, pageSize);
  };
  render() {
    return (
      <Card
        title={"Floating IP Manage Panel" + "(Total: " + this.state.total + ")"}
        extra={
          <>
            <DataFilter
              placeholder="Search..."
              onSearch={(value) => console.log(value)}
              enterButton
            />
            <Button
              style={{ float: "right" }}
              type="primary"
              onClick={this.createFloatingips}
            >
              Create
            </Button>
          </>
        }
      >
        <DataTable
          rowKey="ID"
          columns={this.columns}
          dataSource={this.state.floatingips}
          bordered
          total={this.state.total}
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
export default Floatingips;
