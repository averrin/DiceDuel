$(function(){
    var me = '';
    var c=new WebSocket('ws://localhost:3000/sock');
    c.onopen = function(){
      c.send(JSON.stringify({"Type": "connect", "Message": ""}))
      c.onmessage = function(response){
        var data = JSON.parse(response.data);
        console.log(data);
        switch (data.Type){
            case "connect":
                $("#me").html("Your id: " + data.Message)
                me = data.Message;
                break
            case "new":
                if (data.Message != me){
                    $("#here").append("<li data-id='" + data.Message +"'>" + data.Message + "</li>")
                }
                break
            case "disconnect":
                if (data.Message != me){
                    $("#here li[data-id='"+data.Message+"']").remove()
                }
                break
        }
      }
    }
})