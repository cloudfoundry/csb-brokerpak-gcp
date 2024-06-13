package broadcom.spring.gcp.pubsub.springpubsub.service;

import com.google.cloud.spring.pubsub.core.PubSubTemplate;
import com.google.pubsub.v1.PubsubMessage;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

@Service
public class PubSubService {

    @Value("${spring.cloud.gcp.pubsub.project-id}")
    private String projectId;

    @Autowired
    private PubSubTemplate pubSubTemplate;

    public String publishMessage(String topicName, String message) {
        pubSubTemplate.publish(topicName, message);
        return "Message published";
    }

    public String pullMessages(String subscriptionName) {
        PubsubMessage message = pubSubTemplate.pull(subscriptionName, 1, false).get(0).getPubsubMessage();
        return message.getData().toStringUtf8();
    }
}