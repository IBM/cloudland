import React, { Component } from "react";
import { Card, Table, Button, Popconfirm } from "antd";
import { secrulesListApi } from "../../../api/secrules";
const columns = [
  {
    title: "ID",
    dataIndex: "ID",
    key: "ID",
    width: 80,
    align: "center",
    render: (href) => <a>{href}</a>,
  },
  {
    title: "SecurityGroup",
    dataIndex: "Secgroup",
    render: (href) => <a>{href}</a>,
  },
  {
    title: "RemoteIp",
    dataIndex: "RemoteIp",
  },
  {
    title: "Direction",
    dataIndex: "Direction",
  },
  {
    title: "Protocol",
    dataIndex: "Protocol",
  },
  {
    title: "PortMin|Type",
    dataIndex: "PortMin",
  },
  {
    title: "PortMax|Type",
    dataIndex: "PortMax",
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
class Secrules extends Component {
  constructor(props) {
    super(props);
    console.log("Secgroups.props:", this.props);
    this.state = {
      secrules: [],
      isLoaded: false,
    };
  }
  //组件初始化的时候执行
  componentDidMount() {
    const _this = this;
    console.log("componentDidMount:", this);
    secrulesListApi()
      .then((res) => {
        _this.setState({
          secrules: res.secrules,
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
        title="Security Group Rules Manage Panel"
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
          dataSource={this.state.secrules}
        ></Table>
      </Card>
    );
  }
}
export default Secrules;
