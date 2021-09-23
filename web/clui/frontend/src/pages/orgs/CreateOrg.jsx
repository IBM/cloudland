/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import React, { Component } from "react";
import { withTranslation } from "react-i18next";
import { compose } from "redux";

import { Form, Card, Input, Button, message } from "antd";
import { createOrgApi } from "../../service/orgs";
import "./orgs.css";
const layoutButton = {
  labelCol: { span: 8 },
  wrapperCol: { span: 16 },
};
const layoutForm = {
  labelCol: { span: 6 },
  wrapperCol: { span: 10 },
  LayoutType: "horizontal",
};
class CreateOrg extends Component {
  constructor(props) {
    super(props);
    this.state = {
      currentData: [],
    };
  }

  listOrgs = () => {
    this.props.history.push("/orgs");
  };
  handleSubmit = (e) => {
    e.preventDefault();
    this.props.form.validateFieldsAndScroll((err, values) => {
      if (!err) {
        createOrgApi(values)
          .then((res) => {
            this.props.history.push("/orgs");
          })
          .catch((err) => {
            console.log("handleSubmit-error:", err);
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
        title={t("Create New Organization")}
        extra={
          <Button
            style={{ float: "right" }}
            type="primary"
            onClick={this.listOrgs}
          >
            {t("Return")}
          </Button>
        }
      >
        <Form
          layout={{ ...layoutForm.LayoutType }}
          wrapperCol={{ ...layoutForm.wrapperCol }}
          onSubmit={(e) => this.handleSubmit(e)}
        >
          <Form.Item
            label={t("Organization Name")}
            name="orgname"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("orgname", {
              rules: [
                {
                  required: true,
                },
              ],
            })(<Input />)}
          </Form.Item>
          <Form.Item
            name="owner"
            label={t("Owner")}
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("owner", {})(<Input />)}
          </Form.Item>
          <Form.Item
            wrapperCol={{ ...layoutButton.wrapperCol, offset: 8 }}
            labelCol={{ span: 6 }}
          >
            <Button type="primary" htmlType="submit">
              {t("Create New Organization")}
            </Button>
          </Form.Item>
        </Form>
      </Card>
    );
  }
}
export default compose(
  withTranslation(),
  Form.create({ name: "CreateOrg" })
)(CreateOrg);
