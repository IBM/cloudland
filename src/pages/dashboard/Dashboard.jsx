/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import React, { Component } from "react";
import { Typography, Divider } from "antd";
class Dashboard extends Component {
  render() {
    const { Title, Paragraph } = Typography;

    return (
      <div className="bx--grid content-page">
        <div className="bx--grid content-page-right">
          <Typography className="contentpage-typography">
            <Typography>
              <Title> Brief Introduction</Title>
            </Typography>
            <Title level={4}>1. Create a key</Title>
            <Paragraph className="contentpage-paragraph">
              To create a key, simply input key name and public key into fields
              and submit. If you don't have a key already, use command
              ssh-keygen -f /path/to/your_key to generate one. Most vm instances
              need a key to login.
            </Paragraph>
            <Divider />
            <Title level={4}>2. Create a gateway</Title>
            <Paragraph className="contentpage-paragraph">
              To create a gateway, input a gateway name and select one or
              multiple subnets to attach. The vm instances in the selected
              subnets can access the external network via the gateway, and a
              floating ip can be bound to an instance only after a gateway is
              created.
            </Paragraph>
            <Divider />
            <Title level={4}>3. Launch an instance</Title>
            <Paragraph className="contentpage-paragraph">
              To launch an instance, specify the fields with star including
              hostname, count, image, flavor and primary interface. It is also
              important to select a proper key to login the instance after it
              gets created.
            </Paragraph>
            <Divider />
            <Title level={4}>4. Create a floating IP</Title>
            <Paragraph className="contentpage-paragraph">
              you don't have a key already, use command ssh-keygen -f
              /path/to/your_key to generate one. Most vm instances need a key to
              login.
            </Paragraph>
            <Divider />
            <Title level={4}>
              5. Use your instance and modify security group as needed
            </Title>
            <Paragraph className="contentpage-paragraph">
              If you don't have a key already, use command ssh-keygen -f
              /path/to/your_key to generate one. Most vm instances need a key to
              login.
            </Paragraph>
            <Divider />
            <Title level={4}>6. Create an OpenShift cluster (Advanced)</Title>
            <Paragraph className="contentpage-paragraph">
              <a href="https://github.com/IBM/cloudland/wiki/Manual#create-an-openshift-cluster">
                openshift
              </a>
            </Paragraph>
          </Typography>
        </div>
      </div>
    );
  }
}
export default Dashboard;
