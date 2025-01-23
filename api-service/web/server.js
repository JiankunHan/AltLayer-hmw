const express = require('express');
const app = express();
const port = 8080;

app.use(express.json()); // 解析 JSON 请求体

app.post('/api/data', (req, res) => {
    const inputData = req.body.input;
    console.log('接收到的输入数据:', inputData);

    // 这里可以进行处理，例如：根据输入进行计算或查询数据库
    const result = `你输入的数据是: ${inputData}`;

    res.json({ output: result });
});

app.listen(port, () => {
    console.log(`API 服务器正在监听 http://localhost:${port}`);
});
