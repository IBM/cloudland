/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import React, { Component } from "react";

import { Form, Card, Input, Select, Button, message } from "antd";
import { getUserInforById, editUserInfor } from "../../service/users";
import { withTranslation } from "react-i18next";
import { compose } from "redux";
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
    this.state = {
      isShowEdit: false,
      currentData: [],
      members: [],
    };
    let that = this;
    if (props.match.params.id) {
      getUserInforById(props.match.params.id).then((res) => {
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
    e.preventDefault();
    this.props.form.validateFieldsAndScroll((err, values) => {
      if (!err) {
        console.log("handleSubmit-value:", values);
        editUserInfor(this.props.match.params.id, values).then((res) => {
          this.props.history.push("/users");
        });
      } else {
        message.error(" input wrong information");
      }
    });
  };

  render() {
    const { t } = this.props;
    return (
      <Card
        title={t("Edit_User")}
        extra={
          <Button
            style={{ float: "right" }}
            type="primary"
            onClick={this.listUsers}
          >
            {t("Return")}
          </Button>
        }
      >
        <Form
          layout="horizontal"
          wrapperCol={{ ...layoutForm.wrapperCol }}
          onSubmit={(e) => this.handleSubmit(e)}
        >
          <Form.Item
            label={t("Password")}
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
            label={t("Organizations")}
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
                placeholder={t("Pleaseselect")}
              >
                {this.state.members.map((item, i) => {
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
                {t("Update_User")}
              </Button>
            }
          </Form.Item>
        </Form>
      </Card>
    );
  }
}

export default compose(
  withTranslation(),
  Form.create({ name: "modifyUser" })
)(ModifyUser);
