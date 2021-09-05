/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import React, { Component } from "react";
import { Card, Button, Popconfirm, Input } from "antd";
import { ocpListApi } from "../../service/openshifts";
import DataTable from "../../components/DataTable/DataTable";
const { Search } = Input;
class Openshifts extends Component {
  constructor(props) {
    super(props);
    console.log("Openshifts.props:", this.props);
    this.state = {
      openshifts: [],
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
      key: "ID",
      width: 80,
      align: "center",
    },
    {
      title: "Cluster Name",
      dataIndex: "ClusterName",
      align: "center",
    },
    {
      title: "Base Domain",
      dataIndex: "BaseDomain",
      align: "center",
    },
    {
      title: "Console",
      dataIndex: "console",
      align: "center",
    },
    {
      title: "Version",
      dataIndex: "Version",
      align: "center",
    },
    {
      title: "Flavor",
      dataIndex: "Flavor",
      align: "center",
    },
    {
      title: "HA",
      dataIndex: "Haflag",
      // width: 50,
      align: "center",
    },
    {
      title: "N-Workers",
      dataIndex: "WorkerNum",
      align: "center",
    },
    {
      title: "Status",
      dataIndex: "Status",
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
                console.log("onClick-ocp:", record);
                this.props.history.push("/openshifts/new/" + record.ID);
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
                console.log("confirmed");
                //此处调用api接口进行相关操作
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
    console.log("componentDidMount:", this);
    ocpListApi()
      .then((res) => {
        _this.setState({
          openshifts: res.openshifts,
          filteredList: res.openshifts,
          isLoaded: true,
          total: res.total,
        });
        console.log("openshifts", res);
      })
      .catch((error) => {
        _this.setState({
          isLoaded: false,
          error: error,
        });
      });
  }
  createOpenshift = () => {
    this.props.history.push("/openshifts/new");
  };
  loadData = (page, pageSize) => {
    console.log("loadData~~", page, pageSize);
    const _this = this;
    const offset = (page - 1) * pageSize;
    const limit = pageSize;
    ocpListApi(offset, limit)
      .then((res) => {
        console.log("loadData", res);

        _this.setState({
          openshifts: res.openshifts,
          filteredList: res.openshifts,
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
    ocpListApi(offset, limit)
      .then((res) => {
        console.log("loadData", res);
        _this.setState({
          openshifts: res.openshifts,
          filteredList: res.openshifts,
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
    console.log("event-filter", event.target.value);
    this.getFilteredList(event.target.value);
  };
  getFilteredList = (word) => {
    console.log("getFilteredListr-keyword-ocp", word);
    var keyword = word.toLowerCase();
    if (keyword) {
      this.setState({
        filteredList: this.state.openshifts.filter(
          (item) =>
            item.ID.toString().indexOf(keyword) > -1 ||
            item.BaseDomain.toLowerCase().indexOf(keyword) > -1 ||
            item.ClusterName.toLowerCase().indexOf(keyword) > -1 ||
            item.Flavor.toString().indexOf(keyword) > -1 ||
            item.Haflag.toLowerCase().indexOf(keyword) > -1 ||
            item.Status.toLowerCase().indexOf(keyword) > -1 ||
            item.WorkerNum.toString().indexOf(keyword) > -1 ||
            item.Version.toLowerCase().indexOf(keyword) > -1
        ),
      });

      console.log("filteredList", this.state.filteredList);
    } else {
      this.setState({
        filteredList: this.state.openshifts,
      });
    }
  };
  render() {
    return (
      <Card
        title={
          "Openshift Manage Panel" +
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
                paddingLight: "10px",
              }}
              type="primary"
              onClick={this.createOpenshift}
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
          // scroll={{ y: 600, x: 600 }}
          onPaginationChange={this.onPaginationChange}
          onShowSizeChange={this.onShowSizeChange}
          pageSizeOptions={this.state.pageSizeOptions}
          loading={!this.state.isLoaded}
        />
      </Card>
    );
  }
}

export default Openshifts;
