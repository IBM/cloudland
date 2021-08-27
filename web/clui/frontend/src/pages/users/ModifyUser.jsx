/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import React, { Component } from "react";
import moment from "moment";
import { Form, Card, Input, Select, Button, message } from "antd";
import { getUserInforById, editUserInfor } from "../../service/users";
// import "./users.css";
const layoutButton = {
  labelCol: { span: 8 },
  wrapperCol: { span: 16 },
};
const layoutForm = {
  labelCol: { span: 6 },
  wrapperCol: { span: 10 },
};
class ModifyUser extends Component {
  constructor(props) {
    super(props);
    console.log("ModifyUser~~", props);
    this.state = {
      isShowEdit: false,
      currentData: [],
      members: [],
    };
    let that = this;
    if (props.match.params.id) {
      getUserInforById(props.match.params.id).then((res) => {
        console.log("getUserInforById-res:", res);
        that.setState({
          currentData: res,
          members: res.Members.filter((item) => {
            return item.OrgName;
          }),
          isShowEdit: true,
        });
      });
    }
  }

  listUsers = () => {
    this.props.history.push("/users");
  };
  handleSubmit = (e) => {
    console.log("handleSubmit:", e);
    e.preventDefault();
    this.props.form.validateFieldsAndScroll((err, values) => {
      if (!err) {
        console.log("handleSubmit-value:", values);
        console.log("提交");
        //const _this = this;
        editUserInfor(this.props.match.params.id, values).then((res) => {
          console.log("editUserInfor:", res);

          this.props.history.push("/users");
        });
      } else {
        message.error(" input wrong information");
      }
    });
  };

  render() {
    return (
      <Card
        title={"Edit User"}
        extra={
          <Button type="primary" onClick={this.listUsers}>
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
            label="Password"
            name="password"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("password", {
              rules: [
                {
                  required: true,
                },
              ],
              initialValue: this.state.currentData.password,
            })(<Input />)}
          </Form.Item>
          <Form.Item
            label="Organizations"
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
            })(
              <Select
                mode="tags"
                style={{ width: "100%" }}
                placeholder="Please select"
              >
                {this.state.members.map((item, i) => {
                  console.log("item.OrgName----", item.OrgName);

                  return (
                    <Select.Option key={i} value={item.OrgName}>
                      {item.OrgName}
                    </Select.Option>
                  );
                })}
              </Select>
            )}
          </Form.Item>

          <Form.Item
            wrapperCol={{ ...layoutButton.wrapperCol, offset: 8 }}
            labelCol={{ span: 6 }}
          >
            {
              <Button type="primary" htmlType="submit">
                Update User
              </Button>
            }
          </Form.Item>
        </Form>
      </Card>
    );
  }
}
export default Form.create({ name: "modifyUser" })(ModifyUser);
