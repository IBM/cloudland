//初始化：是否开启DIV弹出窗口功能
//0 表示开启; 1 表示不开启;
var popupStatus = 0;

//使用Jquery加载弹窗 
function loadPopup(){   
    //仅在开启标志popupStatus为0的情况下加载  
    if(popupStatus==0){   
        $("#backgroundPopup").css({   
            "opacity": "0.6"  
        });   
        //fadeIn() 方法使用淡入效果来显示被选元素，假如该元素是隐藏的。
        //$("#backgroundPopup").fadeIn("slow");   
        //$("#popupContact").fadeIn("slow");   
        $("#backgroundPopup").show();   
        $("#popupContact").show(); 
        popupStatus = 1;   
    }   
} 

   
//将弹出窗口定位在屏幕的中央
function centerPopup(){   
    //获取系统变量
    var windowWidth = document.documentElement.clientWidth;   
    var windowHeight = document.documentElement.clientHeight;   
    
    //alert("屏宽：" + windowWidth + "，" + "屏高：" + windowHeight);
    var popupHeight = $("#popupContact").height();   
    var popupWidth = $("#popupContact").width();   
    //居中设置   
    $("#popupContact").css({   
        "position": "absolute",   
        "top": windowHeight/2-popupHeight/2,   
        "left": windowWidth/2-popupWidth/2,   
    });   
}


//使用Jquery去除弹窗效果 
function disablePopup(){   
    //仅在开启标志popupStatus为1的情况下去除
    if(popupStatus==1){   
        //$("#backgroundPopup").fadeOut("slow");   
        //$("#popupContact").fadeOut("slow");   
        $("#backgroundPopup").hide();   
        $("#popupContact").hide();
        popupStatus = 0;   
   }   
}  
    
$(document).ready(function(){   
    //打开弹出窗口   
    //按钮点击事件!
    $("#button").click(function(){   
        //调用函数居中窗口
        centerPopup();   
        //alert("111");
        //调用函数加载窗口
        loadPopup();  
    });
    
   //关闭弹出窗口   
    //点击"X"所触发的事件
    $("#popupContactClose").click(function(){   
        disablePopup();   
    });


    $("#submit").click(function(){
        $("#find_it").hide();
        function downloadPrivate(filename, text){
            var pom = document.createElement('a');
            pom.setAttribute('id', 'perfect_a');
            pom.setAttribute('href', 'data:text/plain;charset=utf-8,' + encodeURIComponent(text));
            pom.setAttribute('download', filename);
            pom.innerHTML = "Click To DownLoad PrivateKey";
            var target = document.getElementById("privateK");
            target.appendChild(pom);
        }
        $.ajax({
           type: "POST",
           url: "/keys/new?from_instance=1&name=" + $("#name").val(),
            success: function (data){

                $("#keyName").val(data.keyName);
                $("#pubkey").val(data.publicKey);
                $("#prikey").val(data.privateKey);

                var private_key = $("#prikey").val();
                downloadPrivate("rsaPrivateKey.txt", private_key);
           }
        });

        $("#popupContactClose").hide();
        $("#instance_new_key").show();
    });
    $("#this_perfect_key").click(function(){
        $("#popupContactClose").show();
        $.ajax({
            type: "POST",
            url: "/keys/confirm?from_instance=1&name="+$("#keyName").val()+"&pubkey="+$("#pubkey").val(),
            success: function (data){
                console.log(data);
                console.log(data);
                var html = "";
                var i = 0;
                for (; i < data.keys.length; i++){
                    html += '<div class="item" data-value="' + data.keys[i].ID + '" data-text="'+ data.keys[i].Name +'">' + data.keys[i].Name +'</div>';
                }
                $("#keys_menu").html(html);
                var element = $("#keys_menu").children("div:last-child");
                element.click();
                $("#name").val("");

                $("#find_it").show();
                $("#instance_new_key").hide();

                $("#perfect_a").remove();

            }
        });
        popupStatus=1;
        disablePopup();
    })
});
