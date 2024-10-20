import mqtt from 'mqtt'

const client = mqtt.connect('mqtt://localhost:8080', {
  reconnectPeriod: 0,
})

client.on("connect", () => {
  console.log("connected");
  client.subscribe("presence", (err) => {
    if (!err) {
      client.publish("presence", "Hello mqtt");
    }
  });
});

client.on("message", (topic, message) => {
  // message is Buffer
  console.log(message.toString());
  client.end();
});


client.on("error", (err) => {
  console.error("error", err);
  client.end();
});
