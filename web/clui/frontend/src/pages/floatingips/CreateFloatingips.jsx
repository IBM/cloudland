import React, { Component } from "react";
import { Form, Card, Button, Select, Input, message } from "antd";
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
  componentWilMount() {
    console.log("componentDidMount:", this);
    instListApi()
      .then((res) => {
        const _this = this;
        console.log("componentDidMount-instances:", res);
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
  handleSubmit = (event) => {
    console.log("handleSubmit-state:", this.state);
    console.log("handleSubmit:", event);
    event.preventDefault();
    this.props.form.validateFieldsAndScroll((err, values) => {
      let _this = this;
      if (!err) {
        // this.setState({
        //   flavors: {
        //     Name: this.state.name,
        //     Cpu: parseInt(this.state.cpu),
        //     Memory: parseInt(this.state.memory),
        //     Disk: parseInt(this.state.disk),
        //     Swap: parseInt(this.state.swap),
        //     Ephemeral: parseInt(this.state.ephemeral),
        //   },
        // });

        let ifaceID = [];
        if (values.privateip != undefined) {
          ifaceID.push(values.privateip);
        }
        if (values.publicip != undefined) {
          ifaceID.push(values.publicip);
        }

        console.log("提交~~~", {
          ftype: values.ftype,
          instance: `${values.instance}`,
          ifaceID: ifaceID,
        });
        createFloatingipApi({
          ftype: values.ftype,
          instance: `${values.instance}`,
          publicip: ifaceID,
        })
          .then((res) => {
            console.log("handleSubmit-res-createFlavorApi:", res);
            this.props.history.push("/flavors");
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
    return (
      <Card
        title={"Create New FloatingIp "}
        extra={
          <Button
            style={{ float: "right" }}
            type="primary"
            onClick={this.listFloatingIps}
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
            label="Instance Address"
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
                  console.log("instance~", item, index);
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
            label="Public IP"
            name="publicip"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("publicip", {
              rules: [],
            })(
              <Input
                name="publicip"
                onChange={(e) => this.setState({ publicip: e.target.value })}
              />
            )}
          </Form.Item>
          <Form.Item
            label="Private IP"
            name="privateip"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("privateip", {
              rules: [],
            })(
              <Input
                name="privateip"
                onChange={(e) => this.setState({ privateip: e.target.value })}
              />
            )}
          </Form.Item>

          <Form.Item
            label="Floating IP Type"
            name="ftype"
            labelCol={{ ...layoutForm.labelCol }}
          >
            {this.props.form.getFieldDecorator("ftype", {
              rules: [],
            })(
              <Select>
                <Select.Option key="public" value="public">
                  public
                </Select.Option>

                <Select.Option key="private" value="private">
                  private
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
                Create FloatingIp
              </Button>
            }
          </Form.Item>
        </Form>
      </Card>
    );
  }
}
export default Form.create({ name: "createFloatingips" })(CreateFloatingips);
