/*

Copyright <holder> All Rights Reserved

SPDX-License-Identifier: Apache-2.0

*/
import React, { Component } from "react";
import { Card } from "antd";
// import * as echarts from "echarts";
import ReactEcharts from "echarts-for-react";
import { getResourceData } from "../../service/dashboard";

import { withTranslation } from "react-i18next";
class Dashboard extends Component {
  constructor(props) {
    super(props);
    this.state = {
      cpu_avail: 0,
      cpu_used: 0,
      disk_avail: 0,
      disk_used: 0,
      mem_avail: 0,
      mem_used: 0,
      prvip_avail: 0,
      prvip_used: 0,
      pubip_avail: 0,
      pubip_used: 0,
      title: "",
      volume_avail: 0,
      volume_used: 0,
    };
  }
  componentDidMount() {
    getResourceData()
      .then((res) => {
        console.log("res-get", res);
        this.setState({
          cpu_avail: res.cpu_avail,
          cpu_used: res.cpu_used,
          disk_avail: res.disk_avail,
          disk_used: res.disk_used,
          mem_avail: res.mem_avail,
          mem_used: res.mem_used,
          prvip_avail: res.prvip_avail,
          prvip_used: res.prvip_used,
          pubip_avail: res.pubip_avail,
          pubip_used: res.pubip_used,
          title: res.title,
          volume_avail: res.volume_avail,
          volume_used: res.volume_used,
        });
      })
      .catch((error) => {
        console.log("error-get", error);
      });
  }
  getOption = () => {
    var labelTop = {
      normal: {
        label: {
          show: true,
          position: "center",
          formatter: "{b}",
          textStyle: {
            baseline: "bottom",
          },
        },
        labelLine: {
          show: false,
        },
      },
    };
    var labelBottom = {
      normal: {
        color: "#ccc",
        label: {
          show: true,
          position: "center",
        },
        labelLine: {
          show: false,
        },
      },
    };
    var labelFromatter = {
      borderRadius: 10,
      borderColor: "#fff",
      borderWidth: 2,
      normal: {
        label: {
          formatter: (params) => {
            console.log("formatter-pa", params);
            let res = params.name + ":" + params.value;
            if (
              params.name.indexOf("volume") > -1 ||
              params.name.indexOf("disk") > -1
            ) {
              res = res + "G";
            } else if (params.name.indexOf("memory") > -1) {
              res = res + "M";
            }
            res = res + "\n" + params.percent + "%";
            return res;
          },
          textStyle: {
            baseline: "top",
          },
          position: "outside",
        },
      },
    };
    var radius = [40, 55];
    let option = {
      title: {
        text: this.state.title,
        subtext: "cpu memory disk volume ip",
        x: "center",
      },
      tooltip: {
        trigger: "item",
        formatter: "{a} <br/>{b} : {c} ({d}%)",
      },
      legend: {
        top: "10%",
        // left: "center",
        data: [
          "cpu_used",
          "memory_used",
          "disk_used",
          "volume_storage_used",
          "public_ip_used",
          "private_ip_used",
        ],
      },

      series: [
        {
          name: "cpu",
          type: "pie",
          center: ["23%", "45%"],
          avoidLabelOverlap: true,
          radius: radius,
          x: "0%", // for funnel
          itemStyle: labelFromatter,

          data: [
            {
              value: this.state.cpu_avail,
              name: "other",
              itemStyle: labelBottom,
            },
            {
              value: this.state.cpu_used,
              name: "cpu_used",
              itemStyle: labelTop,
            },
          ],
        },
        {
          type: "pie",
          center: ["37%", "45%"],
          radius: radius,
          x: "20%", // for funnel
          itemStyle: labelFromatter,
          data: [
            {
              value: this.state.mem_avail,
              name: "other",
              itemStyle: labelBottom,
            },
            {
              value: this.state.mem_used,
              name: "memory_used",
              itemStyle: labelTop,
            },
          ],
        },
        {
          type: "pie",
          center: ["60%", "45%"],
          radius: radius,
          x: "40%", // for funnel
          itemStyle: labelFromatter,
          data: [
            {
              value: this.state.disk_avail,
              name: "other",
              itemStyle: labelBottom,
            },
            {
              value: this.state.disk_used,
              name: "disk_used",
              itemStyle: labelTop,
            },
          ],
        },
        {
          type: "pie",
          center: ["23%", "60%"],
          radius: radius,
          y: "55%", // for funnel
          x: "0%", // for funnel
          itemStyle: labelFromatter,
          data: [
            {
              value: this.state.volume_avail,
              name: "other",
              itemStyle: labelBottom,
            },
            {
              value: this.state.volume_used,
              name: "volume_storage_used",
              itemStyle: labelTop,
            },
          ],
        },
        {
          type: "pie",
          center: ["37%", "60%"],
          radius: radius,
          y: "55%", // for funnel
          x: "20%", // for funnel
          itemStyle: labelFromatter,

          data: [
            {
              value: this.state.pubip_avail,
              name: "other",
              itemStyle: labelBottom,
            },
            {
              value: this.state.pubip_used,
              name: "public_ip_used",
              itemStyle: labelTop,
            },
          ],
        },
        {
          type: "pie",
          center: ["70%", "60%"],
          radius: radius,
          y: "55%", // for funnel
          x: "20%", // for funnel
          itemStyle: labelFromatter,
          data: [
            {
              value: this.state.prvip_avail,
              name: "other",
              itemStyle: labelBottom,
            },
            {
              value: this.state.prvip_used,
              name: "private_ip_used",
              itemStyle: labelTop,
            },
          ],
        },
        {
          name: "cpu",
          type: "pie",
          radius: [0, 25],
          color: "#fff",
          center: ["23%", "45%"],
          itemStyle: labelFromatter,
          label: {
            //set style
            normal: {
              formatter: (params) => {
                return params.name + "\n" + params.value;
              },
              show: true,
              position: "center",
              padding: [0, 0, 20, 0], //it's padding style for word in the middle
              fontSize: 15,
              textStyle: {
                baseline: "top",
                color: "#000",
              },
            },
          },

          data: [
            {
              value: this.state.cpu_avail,
              name: "cpu",
              // itemStyle: labelBottom,
            },
          ],
        },
        {
          name: "memory",
          type: "pie",
          radius: [0, 25],
          x: "20%", // for funnel
          color: "#fff",
          center: ["37%", "45%"],
          itemStyle: labelFromatter,
          label: {
            //set style
            normal: {
              formatter: (params) => {
                return params.name + "\n" + params.value + "M";
              },
              show: true,
              position: "center",
              padding: [0, 0, 20, 0], //it's padding style for word in the middle
              fontSize: 15,
              textStyle: {
                baseline: "top",
                color: "#000",
              },
            },
          },
          data: [{ value: this.state.mem_avail, name: "memory" }],
        },
        {
          name: "disk",
          type: "pie",
          radius: [0, 25],
          color: "#fff",
          x: "40%",
          center: ["60%", "45%"],
          itemStyle: labelFromatter,
          label: {
            //set style
            normal: {
              formatter: (params) => {
                return params.name + "\n" + params.value + "G";
              },
              show: true,
              position: "center",
              padding: [0, 0, 20, 0], //it's padding style for word in the middle
              fontSize: 15,
              textStyle: {
                baseline: "top",
                color: "#000",
              },
            },
          },
          data: [{ value: this.state.disk_avail, name: "disk" }],
        },
        {
          name: "volume",
          type: "pie",
          radius: [0, 25],
          color: "#fff",
          center: ["23%", "60%"],
          y: "55%", // for funnel
          x: "0%", // for funnel
          itemStyle: labelFromatter,
          label: {
            //set style
            normal: {
              formatter: (params) => {
                return params.name + "\n" + params.value + "G";
              },
              show: true,
              position: "center",
              padding: [0, 0, 20, 0], //it's padding style for word in the middle
              fontSize: 15,
              textStyle: {
                baseline: "top",
                color: "#000",
              },
            },
          },
          data: [{ value: this.state.volume_avail, name: "volume" }],
        },
        {
          name: "public_ip",
          type: "pie",
          radius: [0, 25],
          color: "#fff",
          center: ["37%", "60%"],
          y: "55%", // for funnel
          x: "20%", // for funnel
          itemStyle: labelFromatter,
          label: {
            //set style
            normal: {
              formatter: (params) => {
                return params.name + "\n" + params.value;
              },
              show: true,
              position: "center",
              padding: [0, 0, 20, 0], //it's padding style for word in the middle
              fontSize: 15,
              textStyle: {
                baseline: "top",
                color: "#000",
              },
            },
          },
          data: [{ value: this.state.pubip_used, name: "public_ip" }],
        },
        {
          name: "private_ip",
          type: "pie",
          radius: [0, 25],
          color: "#fff",
          center: ["70%", "60%"],
          y: "55%", // for funnel
          x: "20%", // for funnel
          itemStyle: labelFromatter,
          label: {
            //set style
            normal: {
              formatter: (params) => {
                return params.name + "\n" + params.value;
              },
              show: true,
              position: "center",
              padding: [0, 0, 20, 0], //it's padding style for word in the middle
              fontSize: 15,
              textStyle: {
                baseline: "top",
                color: "#000",
              },
            },
          },
          data: [{ value: this.state.prvip_used, name: "private_ip" }],
        },
      ],
    };
    return option;
  };

  render() {
    return (
      <Card className="pie_b">
        <ReactEcharts
          option={this.getOption()}
          style={{ width: "100%", height: "500px" }}
        />
      </Card>
    );
  }
}
export default withTranslation()(Dashboard);
