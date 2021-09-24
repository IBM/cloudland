import React, { Component } from "react";
import { Card, Form, Input, Button, message, Col } from "antd";
import logoLoginImg from "../../assets/img/cland.png";
import "./registers.css";
import { compose } from "redux";
import { withTranslation } from "react-i18next";
import { createUserApi } from "../../service/users";

const layoutForm = {
  labelCol: { span: 8 },
  wrapperCol: { span: 8 },
};
const layoutButton = {
  labelCol: { span: 8 },
  wrapperCol: { span: 16 },
};
class Registers extends Component {
  loginPage = () => {
    this.props.history.push("/login");
  };
  handleSubmit = (e) => {
    e.preventDefault();
    this.props.form.validateFieldsAndScroll((err, values) => {
      if (!err) {
        createUserApi(values)
          .then((res) => {
            console.log("handleSubmit-res-createUserApi:", res);
            this.props.history.push("/users");
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
      <Col span={16} offset={4}>
        <Card
          title={t("Register_User")}
          className="register-form"
          extra={
            <Button
              style={{ float: "right" }}
              type="primary"
              onClick={this.loginPage}
            >
              {t("Return")}
            </Button>
          }
        >
          <div className="login-logo">
            <img src={logoLoginImg} alt="logo" />
            <div className="login-logo-text">
              <h2>CloudLand System</h2>
            </div>
          </div>
          <Form
            layout={{ ...layoutForm.LayoutType }}
            wrapperCol={{ ...layoutForm.wrapperCol }}
            onSubmit={(e) => this.handleSubmit(e)}
          >
            <Form.Item
              label={t("Username")}
              name="username"
              labelCol={{ ...layoutForm.labelCol }}
            >
              {this.props.form.getFieldDecorator("username", {
                rules: [
                  {
                    required: true,
                  },
                ],
              })(<Input />)}
            </Form.Item>
            <Form.Item
              name="password"
              label={t("Password")}
              labelCol={{ ...layoutForm.labelCol }}
            >
              {this.props.form.getFieldDecorator("password", {
                rules: [
                  {
                    required: true,
                  },
                ],
              })(<Input type="password" />)}
            </Form.Item>
            <Form.Item
              name="confirm"
              label={t("Confirm")}
              labelCol={{ ...layoutForm.labelCol }}
            >
              {this.props.form.getFieldDecorator("confirm", {
                rules: [
                  {
                    required: true,
                  },
                ],
              })(<Input type="password" />)}
            </Form.Item>
            <Form.Item
              wrapperCol={{ ...layoutButton.wrapperCol, offset: 10 }}
              labelCol={{ span: 6 }}
            >
              <Button type="primary" htmlType="submit">
                {t("Register")}
              </Button>
            </Form.Item>
          </Form>
        </Card>
      </Col>
    );
  }
}
export default compose(
  withTranslation(),
  Form.create({ name: "registersFrom" })
)(Registers);
