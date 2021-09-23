/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import React, { Component } from "react";
import { Card, Form, Icon, Input, Button, Checkbox, message } from "antd";
import logoLoginImg from "../../assets/img/cland.png";
import "./Login.css";
import { setAll, setToken } from "../../utils/auth";
import { loginApi } from "../../service/auth";
import { compose } from "redux";
import { withTranslation } from "react-i18next";
const layoutForm = {
  labelCol: { span: 8 },
  wrapperCol: { span: 16 },
};

class Login extends Component {
  handleSubmit = (e) => {
    e.preventDefault();
    this.props.form.validateFields((err, values) => {
      if (!err) {
        loginApi({
          username: values.username,
          password: values.password,
        })
          .then((res) => {
            if (res.token) {
              setAll(JSON.stringify(res));
              setToken(res.token);
              message.info(this.props.t("Login_Successfully"));
              this.props.history.push("/");
            } else {
              message.error(this.props.t("Failure_to_Login"));
            }
            console.log(res);
          })

          .catch((err) => {
            // message.error(err.ErrorMsg);
            console.log(err);
          });
      }
    });
  };
  render() {
    const { getFieldDecorator } = this.props.form;
    const { t } = this.props;
    return (
      <Card className="login-form">
        <Form onSubmit={this.handleSubmit}>
          <div className="login-logo">
            <img src={logoLoginImg} alt="logo" />
            <div className="login-logo-text">
              <h2>CloudLand System</h2>
            </div>
          </div>
          <Form.Item name="username" labelCol={{ ...layoutForm.labelCol }}>
            {getFieldDecorator("username", {
              rules: [
                { required: true, message: "Please input your username!" },
              ],
            })(
              <Input
                prefix={
                  <Icon type="user" style={{ color: "rgba(0,0,0,.25)" }} />
                }
                placeholder={t("Username")}
              />
            )}
          </Form.Item>
          <Form.Item name="password" labelCol={{ ...layoutForm.labelCol }}>
            {getFieldDecorator("password", {
              rules: [
                { required: true, message: "Please input your Password!" },
              ],
            })(
              <Input
                prefix={
                  <Icon type="lock" style={{ color: "rgba(0,0,0,.25)" }} />
                }
                type="password"
                placeholder={t("Password")}
              />
            )}
          </Form.Item>
          <Form.Item
          // wrapperCol={{ ...layoutButton.wrapperCol, offset: 6 }}
          // labelCol={{ span: 6 }}
          >
            {getFieldDecorator("remember", {
              valuePropName: "checked",
              initialValue: true,
            })(<Checkbox>{t("Remember_me")}</Checkbox>)}
            <a className="login-form-forgot" href="">
              {t("Forgot_password")}
            </a>
            <Button
              type="primary"
              htmlType="submit"
              className="login-form-button"
            >
              {t("Login")}
            </Button>
            {t("Or")} <a href="">{t("Register now")}</a>
          </Form.Item>
        </Form>
      </Card>
    );
  }
}
export default compose(
  withTranslation(),
  Form.create({ name: "loginFrom" })
)(Login);
