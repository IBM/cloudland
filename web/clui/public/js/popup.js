var popupStatus = 0;

function loadPopup(){   
    if(popupStatus==0){   
        $("#backgroundPopup").css({   
            "opacity": "0.6"  
        });   
  
        $("#backgroundPopup").show();   
        $("#popupContact").show(); 
        popupStatus = 1;   
    }   
} 


function centerPopup(){   

    var windowWidth = document.documentElement.clientWidth;   
    var windowHeight = document.documentElement.clientHeight;   
    
    var popupHeight = $("#popupContact").height();   
    var popupWidth = $("#popupContact").width();   
       
    $("#popupContact").css({   
        "position": "absolute",   
        "top": windowHeight/2-popupHeight/2,   
        "left": windowWidth/2-popupWidth/2,   
    });   
}


function disablePopup(){   
    if(popupStatus==1){   
  
        $("#backgroundPopup").hide();
        $("#name").val("");
        $("#pubkey").val("");
        $("#errorMsg").html("");
        $("#popupContact").hide();
        popupStatus = 0;   
   }   
}  
    
$(document).ready(function(){   

    $("#button").click(function(){   
        centerPopup();   

        loadPopup();  
    });
    

    $("#popupContactClose").click(function(){   
        disablePopup();   
    });


    $("#submit").click(function(){
        function downloadPrivate(filename, text){
            var pom = document.createElement('a');
            pom.setAttribute('id', 'perfect_a');
            pom.setAttribute('href', 'data:text/plain;charset=utf-8,' + encodeURIComponent(text));
            pom.setAttribute('download', filename);
            pom.innerHTML = "Click To DownLoad PrivateKey";
            var target = document.getElementById("privateK");
            target.appendChild(pom);
        }

        if ($("#pubkey").val() !== ""){
            $.ajax({
                type: "POST",
                url: "/keys/new",
                data: {from_instance: "1", name:$("#name").val(), pubkey:$("#pubkey").val()},
                success: function (data){
                    if (data.error !== undefined){
                        console.log("hello world");
                        $("#errorMsg").html(data.error);
                    } else {
                        var html = "";
                        var i = 0;
                        for (; i < data.keys.length; i++) {
                            html += '<div class="item" data-value="' + data.keys[i].ID + '" data-text="' + data.keys[i].Name + '">' + data.keys[i].Name + '</div>';
                        }
                        $("#keys_menu").html(html);
                        var element = $("#keys_menu").children("div:last-child");
                        element.click();
                        $("#name").val("");
                        $("#pubkey").val("");
                        popupStatus=1;
                        disablePopup();
                    }
                }
            });

        }
        else {
            $("#find_it").hide();
            $.ajax({
                type: "POST",
                url: "/keys/new?from_instance=1&name=" + $("#name").val(),
                success: function (data) {
                    console.log(data);

                    $("#keyName").val(data.keyName);
                    $("#pubKey").val(data.publicKey);
                    $("#prikey").val(data.privateKey);

                    var private_key = $("#prikey").val();
                    downloadPrivate("rsaPrivateKey.txt", private_key);
                    $("#popupContactClose").hide();
                    $("#instance_new_key").show();
                }
            });
        }
    });
    $("#this_perfect_key").click(function(){
        $("#popupContactClose").show();
        $.ajax({
            type: "POST",
            url: "/keys/confirm?from_instance=1&name="+$("#keyName").val()+"&pubkey="+$("#pubKey").val(),
            success: function (data){
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
                $("#pubkey").val("");
                $("#errorMsg").html("");

                $("#find_it").show();
                $("#instance_new_key").hide();

                $("#perfect_a").remove();

            }
        });
        popupStatus=1;
        disablePopup();
    })
});