import React, { Component } from "react";
import { Card, Table, Button, Popconfirm } from "antd";
import { floatingipsListApi } from "../../api/floatingips";
const columns = [
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
  },
  {
    title: "InternalIP",
    dataIndex: "IntAddress",
  },
  {
    title: "Instance",
    dataIndex: "Instance.Hostname",
  },
  {
    title: "Zone",
    dataIndex: "Instance.Zone.Name",
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
class Floatingips extends Component {
  constructor(props) {
    super(props);
    this.state = {
      floatingips: [],
      isLoaded: false,
    };
  }
  //组件初始化的时候执行
  componentDidMount() {
    const _this = this;
    console.log("componentDidMount:", this);
    floatingipsListApi()
      .then((res) => {
        _this.setState({
          floatingips: res.floatingips,
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
        title="Floating IP Manage Panel"
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
          dataSource={this.state.floatingips}
        ></Table>
      </Card>
    );
  }
}
export default Floatingips;
