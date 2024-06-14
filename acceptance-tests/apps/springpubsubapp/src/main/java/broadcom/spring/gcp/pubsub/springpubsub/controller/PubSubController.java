package broadcom.spring.gcp.pubsub.springpubsub.controller;

import broadcom.spring.gcp.pubsub.springpubsub.service.PubSubService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/pubsub")
public class PubSubController {
    @Autowired
    private final PubSubService pubSubService;

    @Autowired
    public PubSubController(PubSubService pubSubService) {
        this.pubSubService = pubSubService;
    }

    @GetMapping("/pull-message")
    public String pullMessage(@RequestParam String subscriptionName) {
        return pubSubService.pullMessages(subscriptionName);
    }

    @PostMapping("/post-message")
    public String postMessage(@RequestParam String topicName, @RequestBody String message) {
        return pubSubService.publishMessage(topicName, message);
    }
}
