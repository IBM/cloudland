import React, { Component } from "react";
import { Card, Table, Button, Popconfirm } from "antd";
import { gatewaysListApi } from "../../api/gateways";

const columns = [
  {
    title: "ID",
    dataIndex: "ID",
    key: "ID",
    width: 80,
    align: "center",
  },
  {
    title: "Name",
    dataIndex: "Name",
  },
  {
    title: "Interfaces",
    dataIndex: "Interfaces[0].Address.Address",
  },
  {
    title: "Subnets",
    dataIndex: "Subnets",
  },
  {
    title: "Status",
    dataIndex: "Status",
  },
  {
    title: "Hyper",
    dataIndex: "Hyper + Peer",
  },
  {
    title: "Owner",
    dataIndex: "WorkerNum",
  },
  {
    title: "Zone",
    dataIndex: "Zone.Name",
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
class Gateways extends Component {
  constructor(props) {
    super(props);
    console.log("gateways.props:", this.props);
    this.state = {
      gateways: [],
      isLoaded: false,
    };
  }
  //组件初始化的时候执行
  componentDidMount() {
    const _this = this;
    //const hyper =''
    console.log("componentDidMount:", this);
    gatewaysListApi()
      .then((res) => {
        _this.setState({
          gateways: res.gateways,
          isLoaded: true,
        });
        //hyper = this.state.gateways.Interfaces[0].Hyper + Interfaces[2].Hyper
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
        title="Gateway Manage Panel"
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
          dataSource={this.state.gateways}
        ></Table>
      </Card>
    );
  }
}
export default Gateways;
