var myChart = echarts.init(document.getElementById('dashboard'));
var labelTop = {
    normal : {
        label : {
            show : true,
            position : 'center',
            formatter : '{b}',
            textStyle: {
                baseline : 'bottom'
            }
        },
        labelLine : {
            show : false
        }
    }
};
var labelFromatter = {
    normal : {
        label : {
            formatter : function (params){
                res = params.value
                if (params.name.indexOf('volume') > -1 || params.name.indexOf('disk') > -1) {
                    res = res + 'G'
                } else if (params.name.indexOf('memory') > -1) {
                    res = res + 'M'
                }
                res = res + '\n' + params.percent + '%'
                return res
            },
            textStyle: {
                baseline : 'top'
            },
            position: 'outside'
        }
    },
}
var labelBottom = {
    normal : {
        color: '#ccc',
        label : {
            show : true,
            position : 'center'
        },
        labelLine : {
            show : false
        }
    }
};
var radius = [40, 55];
option = {
    title: {
        subtext : '处理器 内存 磁盘 卷 IP',
        x : 'center',
    },
    legend: {
        x : 'center',
        y : 'center',
        data:[
            '处理器使用','内存使用','磁盘使用','卷存储使用','公网IP使用','私网IP使用'
        ]
    },
    toolbox: {
        show : true,
        feature : {
            dataView : {show: true, readOnly: false},
            magicType : {
                show: true, 
                type: ['pie', 'funnel'],
                option: {
                    funnel: {
                        width: '20%',
                        height: '30%',
                        itemStyle : {
                            normal : {
                                label : {
                                    formatter : function (params){
                                        return 'other\n' + params.value + '%\n'
                                    },
                                    textStyle: {
                                        baseline : 'middle'
                                    }
                                }
                            },
                        } 
                    }
                }
            },
            restore : {show: true},
            saveAsImage : {show: true}
        }
    },
    series : [
        {
            name: 'cpu',
            type : 'pie',
            center : ['20%', '30%'],
            avoidLabelOverlap: true,
            radius : radius,
            x: '0%', // for funnel
            name: 'cpu',
            itemStyle : labelFromatter,
            data : [
                {name:'other', itemStyle : labelBottom},
                {name:'处理器使用', itemStyle : labelTop}
            ]
        },
        {
            type : 'pie',
            center : ['50%', '30%'],
            radius : radius,
            x:'20%', // for funnel
            itemStyle : labelFromatter,
            data : [
                {name:'other', itemStyle : labelBottom},
                {name:'内存使用', itemStyle : labelTop}
            ]
        },
        {
            type : 'pie',
            center : ['80%', '30%'],
            radius : radius,
            x:'40%', // for funnel
            itemStyle : labelFromatter,
            data : [
                {name:'other', itemStyle : labelBottom},
                {name:'磁盘使用', itemStyle : labelTop}
            ]
        },
        {
            type : 'pie',
            center : ['20%', '75%'],
            radius : radius,
            y: '55%',   // for funnel
            x: '0%',    // for funnel
            itemStyle : labelFromatter,
            data : [
                {name:'other', itemStyle : labelBottom},
                {name:'卷存储使用', itemStyle : labelTop}
            ]
        },
        {
            type : 'pie',
            center : ['50%', '75%'],
            radius : radius,
            y: '55%',   // for funnel
            x:'20%',    // for funnel
            itemStyle : labelFromatter,
            data : [
                {name:'other', itemStyle : labelBottom},
                {name:'公网IP使用', itemStyle : labelTop}
            ]
        },
        {
            type : 'pie',
            center : ['80%', '75%'],
            radius : radius,
            y: '55%',   // for funnel
            x:'20%',    // for funnel
            itemStyle : labelFromatter,
            data : [
                {name:'other', itemStyle : labelBottom},
                {name:'私网IP使用', itemStyle : labelTop}
            ]
        },
{
	    name: 'cpu',
            type:'pie',
            radius: [0, 25],
            color: '#fff',
            center : ['20%', '30%'],
            label: {
                normal: {
            formatter : function (params){
//		console.log(params)
                return params.name + '\n' + params.value
            },
            textStyle: {
                baseline : 'top',
                color: '#000'
            },
		    show: true,
                    position: 'center'
                }
            },
            data:[
                {name:'处理器'},
            ]
        },
{
	    name: 'memory',
            type:'pie',
            radius: [0, 25],
            color: '#fff',
            center : ['50%', '30%'],
            label: {
                normal: {
            formatter : function (params){
//		console.log(params)
                return params.name + '\n' + params.value + 'M'
            },
            textStyle: {
                baseline : 'top',
                color: '#000'
            },
		    show: true,
                    position: 'center'
                }
            },
            data:[
                {name:'内存'},
            ]
        },
{
	    name: 'disk',
            type:'pie',
            radius: [0, 25],
            color: '#fff',
            center : ['80%', '30%'],
            label: {
                normal: {
            formatter : function (params){
//		console.log(params)
                return params.name + '\n' + params.value + 'G'
            },
            textStyle: {
                baseline : 'top',
                color: '#000'
            },
		    show: true,
                    position: 'center'
                }
            },
            data:[
                {name:'磁盘'},
            ]
        },
{
	    name: 'volume',
            type:'pie',
            radius: [0, 25],
            color: '#fff',
            center : ['20%', '75%'],
            label: {
                normal: {
            formatter : function (params){
//		console.log(params)
                return params.name + '\n' + params.value + 'G'
            },
            textStyle: {
                baseline : 'top',
                color: '#000'
            },
		    show: true,
                    position: 'center'
                }
            },
            data:[
                {name:'卷存储'},
            ]
        },
{
	    name: 'public_ip',
            type:'pie',
            radius: [0, 25],
            color: '#fff',
            center : ['50%', '75%'],
            label: {
                normal: {
            formatter : function (params){
//		console.log(params)
                return params.name + '\n' + params.value
            },
            textStyle: {
                baseline : 'top',
                color: '#000'
            },
		    show: true,
                    position: 'center'
                }
            },
            data:[
                {name:'公网IP'},
            ]
        },
{
	    name: 'private_ip',
            type:'pie',
            radius: [0, 25],
            color: '#fff',
            center : ['80%', '75%'],
            label: {
                normal: {
            formatter : function (params){
//		console.log(params)
                return params.name + '\n' + params.value
            },
            textStyle: {
                baseline : 'top',
                color: '#000'
            },
		    show: true,
                    position: 'center'
                }
            },
            data:[
                {name:'私网IP'},
            ]
        },
    ]
};
  
getResourceData();

var int=self.setInterval("getResourceData()",10000);

function getResourceData()
{
$.ajax({
    url: "/dashboard/getdata",
    type: 'GET',
    success: function (data) {
        if (data) {
                if (data.title == "System Resource Usage Ratio") {
                        option.title.text = "系统资源使用率"
                } else {
                        option.title.text = "组织配额使用率"
                }
                option.series[0].data[0].value = data.cpu_avail
                option.series[0].data[1].value = data.cpu_used
                option.series[1].data[0].value = data.mem_avail
                option.series[1].data[1].value = data.mem_used
                option.series[2].data[0].value = data.disk_avail
                option.series[2].data[1].value = data.disk_used
                option.series[3].data[0].value = data.volume_avail
                option.series[3].data[1].value = data.volume_used
                option.series[4].data[0].value = data.pubip_avail
                option.series[4].data[1].value = data.pubip_used
                option.series[5].data[0].value = data.prvip_avail
                option.series[5].data[1].value = data.prvip_used
                option.series[6].data[0].value = data.cpu_avail + data.cpu_used
                option.series[7].data[0].value = data.mem_avail + data.mem_used
                option.series[8].data[0].value = data.disk_avail + data.disk_used
                option.series[9].data[0].value = data.volume_avail + data.volume_used
                option.series[10].data[0].value = data.pubip_avail + data.pubip_used
                option.series[11].data[0].value = data.prvip_avail + data.prvip_used
                myChart.setOption(option);
        }
    },
    error: function (jqXHR, textStatus, errorThrown) {
        window.location.href = "/error?ErrorMsg=" + jqXHR.responseText;
    }
});
}





