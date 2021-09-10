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
      owerUser: [],
      members: [],
    };
    let that = this;
    if (props.match.params.id) {
      getOrgInforById(props.match.params.id).then((res) => {
        console.log("getOrgInforById-res:", res);
        that.setState({
          currentData: res,
          owerUser: res.OwnerUser,
          members: res.Members.filter((item) => {
            return { UserName: item.UserName, Role: item.Role };
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
          <Button
            style={{ float: "right" }}
            type="primary"
            onClick={this.listOrgs}
          >
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
              initialValue: this.state.currentData.name,
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
              initialValue: this.state.owerUser.username,
            })(<Input />)}
          </Form.Item>
          <Form.Item
            label="Member List"
            style={{ "font-size": "20px", "font-weight": "bold" }}
          ></Form.Item>
          {this.state.members.map((item, index) => {
            return [
              <Form.Item
                label=""
                name="names"
                // labelCol={{ ...layoutForm.labelCol, offset: 6 }}
                wrapperCol={{ ...layoutForm.wrapperCol, offset: 6 }}
              >
                {this.props.form.getFieldDecorator("names", {
                  rules: [
                    {
                      required: true,
                    },
                  ],
                  initialValue: item.UserName,
                })(<Input />)}
              </Form.Item>,
              <Form.Item
                label=""
                name="roles"
                wrapperCol={{ ...layoutForm.wrapperCol, offset: 6 }}
              >
                {this.props.form.getFieldDecorator("roles", {
                  rules: [
                    {
                      required: true,
                    },
                  ],
                  initialValue: item.Role.toString(),
                })(
                  <Select>
                    <Select.Option value="0">None</Select.Option>
                    <Select.Option value="1">Reader</Select.Option>
                    <Select.Option value="2">Writer</Select.Option>
                    <Select.Option value="3">Owner</Select.Option>
                    <Select.Option value="4">Admin</Select.Option>
                  </Select>
                )}
              </Form.Item>,
            ];
          })}
          <Form.Item
            wrapperCol={{ ...layoutButton.wrapperCol, offset: 8 }}
            labelCol={{ span: 6 }}
          >
            {
              <Button type="primary" htmlType="submit">
                Update Organization
              </Button>
            }
          </Form.Item>
        </Form>
      </Card>
    );
  }
}
export default Form.create({ name: "modifyOrg" })(ModifyOrg);
