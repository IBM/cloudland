import React, { Component } from "react";
import { Card, Table, Button, Popconfirm } from "antd";
import { ocpListApi } from "../../api/openshifts";
const columns = [
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
  },
  {
    title: "Base Domain",
    dataIndex: "BaseDomain",
  },
  {
    title: "Console",
    dataIndex: "console",
  },
  {
    title: "Version",
    dataIndex: "Version",
  },
  {
    title: "Flavor",
    dataIndex: "Flavor",
  },
  {
    title: "HA",
    dataIndex: "Haflag",
  },
  {
    title: "N-Workers",
    dataIndex: "WorkerNum",
  },
  {
    title: "Status",
    dataIndex: "Status",
  },
  {
    title: "Action",
    render: (txt, record, index) => {
      return (
        <div>
          <Button type="primary" size="small">
            Edit
          </Button>
          <Popconfirm
            title="确定删除此项?"
            onCancel={() => {
              console.log("用户取消删除");
            }}
            onConfirm={() => {
              console.log("用户确认删除");
              //此处调用api接口进行相关操作
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
class Openshifts extends Component {
  constructor(props) {
    super(props);
    console.log("Openshifts.props:", this.props);
    this.state = {
      openshifts: [],
      isLoaded: false,
    };
  }
  //组件初始化的时候执行
  componentDidMount() {
    const _this = this;
    console.log("componentDidMount:", this);
    ocpListApi()
      .then((res) => {
        _this.setState({
          openshifts: res.openshifts,
          isLoaded: true,
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
  render() {
    return (
      <Card
        title="Openshift Manage Panel"
        extra={
          <Button type="primary" size="small">
            Create
          </Button>
        }
      >
        <Table
          rowKey="ID"
          columns={columns}
          bordered
          dataSource={this.state.openshifts}
        ></Table>
      </Card>
    );
  }
}
export default Openshifts;
