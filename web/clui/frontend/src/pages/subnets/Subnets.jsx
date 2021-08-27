/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import React, { Component } from "react";
import { Card, Button, Popconfirm, message } from "antd";
import { subnetsListApi, delSubInfor } from "../../service/subnets";
import DataTable from "../../components/DataTable/DataTable";
import DataFilter from "../../components/Filter/DataFilter";

class Subnets extends Component {
  constructor(props) {
    super(props);
    this.state = {
      subnets: [],
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
      //render: (txt, record, index) => index + 1,
    },
    {
      title: "Name",
      dataIndex: "Name",
      align: "center",
    },
    {
      title: "Network",
      dataIndex: "Network",
      align: "center",
    },
    {
      title: "Netmask",
      dataIndex: "Netmask",
      align: "center",
    },
    {
      title: "Zones",
      dataIndex: "Zones",
      align: "center",
      render: (Zones) => (
        <span>
          {Zones.map((zones) => {
            return zones.Name;
          })}
        </span>
      ),
    },
    {
      title: "Vlan",
      dataIndex: "Vlan",
      align: "center",
    },
    {
      title: "Hyper",
      dataIndex: "Netlink.Hyper",
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
                this.props.history.push("/subnets/new/" + record.ID);
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
                delSubInfor(record.ID)
                  .then((res) => {
                    //const _this = this;
                    console.log("delSubInfor-res", res);
                    message.success(res.Msg);
                    this.loadData(this.state.current, this.state.pageSize);

                    console.log("用户~~", res);
                    console.log("用户~~state", this.state);
                  })
                  .catch((err) => {
                    console.log("用户~~err", err);
                    // message.error(err.response.data.ErrorMsg);
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
  componentDidMount() {
    const _this = this;
    subnetsListApi()
      .then((res) => {
        console.log("componentDidMount-orgsListApi:", res);
        _this.setState({
          subnets: res.subnets,
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
    console.log("loadData~~", page, pageSize);
    const _this = this;
    const offset = (page - 1) * pageSize;
    const limit = pageSize;
    subnetsListApi(offset, limit)
      .then((res) => {
        console.log("loadData", res);

        _this.setState({
          subnets: res.subnets,
          isLoaded: true,
          total: res.total,
          pageSize: limit,
          current: page,
        });
        console.log("loadData-page-", page, _this.state);
      })
      .catch((error) => {
        message.error(error.response.data.ErrorMsg);
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
    subnetsListApi(offset, limit)
      .then((res) => {
        console.log("loadData", res);
        _this.setState({
          subnets: res.subnets,
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
  createSubnets = () => {
    this.props.history.push("/subnets/new");
  };

  render() {
    return (
      <Card
        title={"Subnet Manage Panel" + "(Total: " + this.state.total + ")"}
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
              onClick={this.createSubnets}
            >
              Create
            </Button>
          </>
        }
      >
        <DataTable
          rowKey="ID"
          columns={this.columns}
          dataSource={this.state.subnets}
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
export default Subnets;
