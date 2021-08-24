/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import React, { Component } from "react";
import { Card, Button, Popconfirm, message } from "antd";
import { Link } from "react-router-dom";
import { secgroupsListApi, delSecgroupInfor } from "../../service/secgroups";
import DataTable from "../../components/DataTable/DataTable";

class Secgroups extends Component {
  constructor(props) {
    super(props);
    console.log("Secgroups.props:", this.props);
    this.state = {
      secgroups: [],
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
      render: (record) => (
        <Link
          to={`/secgroups/${record}/secrules`}
          // to={{
          //   pathname: `/secgroups/${record}/secrules`,
          //   state: { sgID: record },
          // }}
        >
          {record}
        </Link>
      ),
    },
    {
      title: "Name",
      dataIndex: "Name",
      align: "center",
    },
    {
      title: "IsDefault",
      dataIndex: "IsDefault",
      align: "center",
    },
    {
      title: "Owner",
      dataIndex: "OwnerInfo.name",
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
                this.props.history.push("/secgroups/new/" + record.ID);
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
                delSecgroupInfor(record.ID).then((res) => {
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
  componentWillMount() {
    const _this = this;
    console.log("componentDidMount:", this);
    secgroupsListApi()
      .then((res) => {
        _this.setState({
          secgroups: res.secgroups,
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
  createSecgroups = () => {
    this.props.history.push("/secgroups/new");
  };
  loadData = (page, pageSize) => {
    console.log("loadData~~", page, pageSize);
    const _this = this;
    const offset = (page - 1) * pageSize;
    const limit = pageSize;
    secgroupsListApi(offset, limit)
      .then((res) => {
        console.log("loadData", res);

        _this.setState({
          secgroups: res.secgroups,
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
    secgroupsListApi(offset, limit)
      .then((res) => {
        console.log("loadData", res);
        _this.setState({
          secgroups: res.secgroups,
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
  render() {
    return (
      <Card
        title={
          "Security Group Manage Panel" + "(Total: " + this.state.total + ")"
        }
        extra={
          <Button type="primary" size="small" onClick={this.createSecgroups}>
            Create
          </Button>
        }
      >
        <DataTable
          rowKey="ID"
          columns={this.columns}
          dataSource={this.state.secgroups}
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
export default Secgroups;
