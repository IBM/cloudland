/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import React, { Component } from "react";
import { Form, Card, Input, Select, Button, message } from "antd";
import { getOrgInforById, editOrgInfor } from "../../service/orgs";
import "./orgs.css";
const layoutButton = {
  labelCol: { span: 8 },
  wrapperCol: { span: 16 },
};
const layoutForm = {
  labelCol: { span: 6 },
  wrapperCol: { span: 10 },
};
class ModifyOrg extends Component {
  constructor(props) {
    super(props);
    this.state = {
      isShowEdit: false,
      currentData: [],
      member: [],

    };
    let that = this;
    if (props.match.params.id) {
      getOrgInforById(props.match.params.id).then((res) => {
        console.log("getOrgInforById-res:", res);
        that.setState({
          currentData: res,
          members: res.Members.filter((item) => {
            return item.Username;
          }),          
          isShowEdit: true,
        });
      });
    }
  }

  listOrgs = () => {
    this.props.history.push("/orgs");
  };
  handleSubmit = (e) => {
    console.log("handleSubmit:", e);
    e.preventDefault();
    this.props.form.validateFieldsAndScroll((err, values) => {
      if (!err) {
        console.log("handleSubmit-value:", values);
        //const _this = this;
        editOrgInfor(this.props.match.params.id, values).then((res) => {

          this.props.history.push("/orgs");
        });
      } else {
        message.error(" input wrong information");
      }
    });
  };

  render() {
    return (
      <Card
        title={"Edit Organization"}
        extra={
          <Button type="primary" onClick={this.listOrgs}>
            Return
          </Button>
        }
      >
        <Form
          layout="horizontal"
          wrapperCol={{ ...layoutForm.wrapperCol }}
          onSubmit={(e) => this.handleSubmit(e)}
        >
          <Form.Item
            label="Organization Name"
            name="orgname"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("orgname", {
              rules: [
                {
                  required: true,
                },
              ],
              initialValue: this.state.currentData.members.Username,
            })(<Input />)}
          </Form.Item>
          <Form.Item
            label="Owner"
            name="owner"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("owner", {
              rules: [
                {
                  required: true,
                },
              ],
              initialValue: this.state.currentData.OwnerUser.Username,
            })(<Input />)}
          </Form.Item>
          <Form.Item
            label="Members"
            name="members"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("members", {
              rules: [
                {
                  required: true,
                },
              ],
              initialValue: this.state.members.map((item) => {
                return item.OrgName;
              }),
            })(<Input />)}
          </Form.Item>
          <Form.Item
            wrapperCol={{ ...layoutButton.wrapperCol, offset: 8 }}
            labelCol={{ span: 6 }}
          >
            {
              <Button type="primary" htmlType="submit">
                Update Registry
              </Button>
            }
          </Form.Item>
        </Form>
      </Card>
    );
  }
}
export default Form.create({ name: "modifyUser" })(ModifyUser);
