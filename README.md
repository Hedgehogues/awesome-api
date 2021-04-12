# Awesome-api

I noticed that I was doing the same job of setting up and creating the client every time. Therefore, I decided to 
simplify my life a little and make a wrapper over the standard http client, which will allow you to write more 
high-level code.

This is a general approach that can be implemented in any programming language. He professes a declarative description 
of API methods. This means that you can describe each method in your API independently of the method calls. In this 
case, you will use a single client that is created for a specific api host. Each method must have only some fields with 
data. It is enough.

# Example

We show simple example. Artificial example. Client definition:

    apiClient := cl.NewClient("https", "your.domain", nil, nil)
    var m cl.Method
    
Request to the first API method:
    
    m := methods.NewYourMethod1(&params1)
	resp, err := api.Request(m, nil)
	fmt.Println("Method 1", resp.Map())
    
Request to the second API method:

    m := methods.NewYourMethod2(&params2)
	resp, err := api.Request(m, nil)
	fmt.Println("Method 2", resp.Map())
    
Example of response:
       
    > Text: {"id":7,"url":"https://i.thatcopy.pw/cat/b9QYp8T.jpg","webpurl":"https://i.thatcopy.pw/cat-webp/b9QYp8T.webp","x":73.02,"y":38.44}
    > Bytes: [123 34 105 100 34 58 55 44 34 117 114 108 34 58 34 104 116 116 112 115 58 47 47 105 46 116 104 97 116 99 111 112 121 46 112 119 47 99 97 116 47 98 57 81 89 112 56 84 46 106 112 103 34 44 34 119 101 98 112 117 114 108 34 58 34 104 116 116 112 115 58 47 47 105 46 116 104 97 116 99 111 112 121 46 112 119 47 99 97 116 45 119 101 98 112 47 98 57 81 89 112 56 84 46 119 101 98 112 34 44 34 120 34 58 55 51 46 48 50 44 34 121 34 58 51 56 46 52 52 125]
    > Map: map[id:7 url:https://i.thatcopy.pw/cat/b9QYp8T.jpg webpurl:https://i.thatcopy.pw/cat-webp/b9QYp8T.webp x:73.02 y:38.44]
    > Struct: {7 https://i.thatcopy.pw/cat/b9QYp8T.jpg https://i.thatcopy.pw/cat-webp/b9QYp8T.webp 73.02 38.44}
    
# Appendix
    
Real example, you can find [here](example/main.go).
