/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import React, { Component } from "react";
import { Typography, Divider } from "antd";
import { withTranslation } from "react-i18next";
import { Link } from "react-router-dom";
class Dashboard extends Component {
  render() {
    const { Title, Paragraph } = Typography;
    const { t } = this.props;
    return (
      <div className="bx--grid content-page">
        <div className="bx--grid content-page-right">
          <Typography className="contentpage-typography">
            <Typography>
              <Title>{t("Brief_Instructions")}</Title>
            </Typography>
            <Title level={4}>1. {t("Create_a_key")}</Title>
            <Paragraph className="contentpage-paragraph">
              {t("Create_a_key_content")}
            </Paragraph>
            <Divider />
            <Title level={4}>2. {t("Create_a_gateway")}</Title>
            <Paragraph className="contentpage-paragraph">
              {t("Create_a_gateway_content")}
            </Paragraph>
            <Divider />
            <Title level={4}>3. {t("Launch_an_instance")}</Title>
            <Paragraph className="contentpage-paragraph">
              {t("Launch_an_instance_content")}
            </Paragraph>
            <Divider />
            <Title level={4}>4. {t("Create_a_floating_IP")}</Title>
            <Paragraph className="contentpage-paragraph">
              {t("Create_a_floating_IP_content")}
            </Paragraph>
            <Divider />
            <Title level={4}>
              5. {t("Use_your_instance_and_modify_security_group_as_needed")}
            </Title>
            <Paragraph className="contentpage-paragraph">
              {t("ssh_key_login_cmd")}
              {<br />}
              {t("Use_your_instance_content")}
            </Paragraph>
            <Divider />
            <Title level={4}>
              6. {t("Create_an_OpenShift_cluster")}({t("Advanced")})
            </Title>
            <Paragraph className="contentpage-paragraph">
              <Link
                to={
                  "https://github.com/IBM/cloudland/wiki/Manual#create-an-openshift-cluster"
                }
              >
                https://github.com/IBM/cloudland/wiki/Manual#create-an-openshift-cluster
              </Link>
            </Paragraph>
          </Typography>
        </div>
      </div>
    );
  }
}
export default withTranslation()(Dashboard);
