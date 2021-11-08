import React, { Component } from "react";
import { Form, Card, Button, Select, message } from "antd";
import { withTranslation } from "react-i18next";
import { compose } from "redux";
import { createFloatingipApi } from "../../service/floatingips";
import { instListApi } from "../../service/instances";

const layoutButton = {
  labelCol: { span: 8 },
  wrapperCol: { span: 16 },
};
const layoutForm = {
  labelCol: { span: 6 },
  wrapperCol: { span: 10 },
};
class CreateFloatingips extends Component {
  constructor(props) {
    super(props);
    this.state = {
      instances: [],
      publicip: "",
      privateip: "",
      ftype: [],
      instance: "",
    };
  }
  //it will executed while initting component
  componentDidMount() {
    instListApi()
      .then((res) => {
        const _this = this;
        _this.setState({
          instances: res.instances,
          isLoaded: true,
          pagination: {
            total: res.total,
          },
        });
      })
      .catch((error) => {
        const _this = this;
        _this.setState({
          isLoaded: false,
          error: error,
        });
      });
  }
  listFloatingIps = () => {
    this.props.history.push("/floatingips");
  };
  //submit form
  handleSubmit = (event) => {
    event.preventDefault();
    this.props.form.validateFieldsAndScroll((err, values) => {
      if (!err) {
        let ifaceID = [];
        if (values.privateip !== undefined) {
          ifaceID.push(values.privateip);
        }
        if (values.publicip !== undefined) {
          ifaceID.push(values.publicip);
        }

        createFloatingipApi({
          ftype: values.ftype,
          instance: `${values.instance}`,
          publicip: ifaceID,
        })
          .then((res) => {
            this.props.history.push("/floatingips");
          })
          .catch((err) => {
            console.log("Error,Create floating IP handleSubmit-error:", err);
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
        title={t("Create New Floating Ip")}
        extra={
          <Button
            style={{ float: "right" }}
            type="primary"
            onClick={this.listFloatingIps}
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
            label={t("Instance Address")}
            name="instance"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("instance", {
              rules: [
                {
                  required: true,
                },
              ],
            })(
              <Select>
                {this.state.instances.map((item, index) => {
                  return (
                    <Select.Option key={index} value={item.ID}>
                      {item.ID} - {item.Hostname}-
                      {item.Interfaces.map((val) => {
                        return val.Address.Address;
                      })}
                    </Select.Option>
                  );
                })}
              </Select>
            )}
          </Form.Item>

          <Form.Item
            label={t("Floating IP type")}
            name="ftype"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("ftype", {
              rules: [],
            })(
              <Select placeholder={t("Type")}>
                <Select.Option key="public" value="public">
                  {t("public")}
                </Select.Option>

                <Select.Option key="private" value="private">
                  {t("private")}
                </Select.Option>
              </Select>
            )}
          </Form.Item>

          <Form.Item
            wrapperCol={{ ...layoutButton.wrapperCol, offset: 8 }}
            labelCol={{ span: 6 }}
          >
            {
              <Button type="primary" htmlType="submit">
                {t("Create New Floating Ip")}
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
  Form.create({ name: "createFloatingips" })
)(CreateFloatingips);
