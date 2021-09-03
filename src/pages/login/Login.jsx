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
class Login extends Component {
  handleSubmit = (e) => {
    e.preventDefault();
    this.props.form.validateFields((err, values) => {
      if (!err) {
        console.log("Received values of form: ", values);
        // setToken(values.username);
        // this.props.history.push("/");
        loginApi({
          username: values.username,
          password: values.password,
        })
          .then((res) => {
            if (res.token) {
              console.log("login-res:", res);
              setAll(JSON.stringify(res));
              setToken(res.token);
              message.info("Login Successfully");
              this.props.history.push("/");
            } else {
              //message.info(res.ErrorMsg);
              message.error("Failure to Login");
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
    return (
      <Card className="login-form">
        <Form onSubmit={this.handleSubmit}>
          <div className="login-logo">
            <img src={logoLoginImg} alt="logo" />
            <div className="login-logo-text">
              <h2>CloudLand System</h2>
            </div>
          </div>
          <Form.Item>
            {getFieldDecorator("username", {
              rules: [
                { required: true, message: "Please input your username!" },
              ],
            })(
              <Input
                prefix={
                  <Icon type="user" style={{ color: "rgba(0,0,0,.25)" }} />
                }
                placeholder="Username"
              />
            )}
          </Form.Item>
          <Form.Item>
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
                placeholder="Password"
              />
            )}
          </Form.Item>
          <Form.Item>
            {getFieldDecorator("remember", {
              valuePropName: "checked",
              initialValue: true,
            })(<Checkbox>Remember me</Checkbox>)}
            <a className="login-form-forgot" href="">
              Forgot password
            </a>
            <Button
              type="primary"
              htmlType="submit"
              className="login-form-button"
            >
              Log in
            </Button>
            Or <a href="">register now!</a>
          </Form.Item>
        </Form>
      </Card>
    );
  }
}
export default Form.create({ name: "loginFrom" })(Login);
