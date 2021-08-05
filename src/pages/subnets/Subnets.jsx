import React, { Component } from "react";
import { Card, Table, Button, Popconfirm } from "antd";
import { subnetsListApi } from "../../api/subnets";
const columns = [
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
  },
  {
    title: "Network",
    dataIndex: "Network",
  },
  {
    title: "Netmask",
    dataIndex: "Netmask",
  },
  {
    title: "Zones",
    dataIndex: "Zones[0].Name",
  },
  {
    title: "Vlan",
    dataIndex: "Vlan",
  },
  {
    title: "Hyper",
    dataIndex: "Netlink.Hyper",
  },
  {
    title: "Owner",
    dataIndex: "OwnerInfo.name",
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
class Subnets extends Component {
  constructor(props) {
    super(props);
    this.state = {
      subnets: [],
      isLoaded: false,
    };
  }
  componentDidMount() {
    const _this = this;
    subnetsListApi()
      .then((res) => {
        console.log("componentDidMount-orgsListApi:", res);
        _this.setState({
          subnets: res.subnets,
          isLoaded: true,
        });
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
        title="Subnet Manage Panel"
        extra={
          <Button
            type="primary"
            size="small"
            //onClick={() => this.props.history.push("`/orgs/${orgsid}`")}
          >
            Create
          </Button>
        }
      >
        <Table
          rowKey="ID"
          columns={columns}
          bordered
          dataSource={this.state.subnets}
        ></Table>
      </Card>
    );
  }
}
export default Subnets;
