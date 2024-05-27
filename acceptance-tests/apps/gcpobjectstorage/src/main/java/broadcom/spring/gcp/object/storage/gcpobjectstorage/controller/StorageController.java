package broadcom.spring.gcp.object.storage.gcpobjectstorage.controller;

import broadcom.spring.gcp.object.storage.gcpobjectstorage.service.StorageService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/storage")
public class StorageController {

    @Autowired
    private final StorageService storageService;

    @Autowired
    public StorageController(StorageService storageService) {
        this.storageService = storageService;
    }

    @GetMapping("/read")
    public String readObject(@RequestParam String bucketName, @RequestParam String objectName) {
        return storageService.readObject(bucketName, objectName);
    }

    @PostMapping("/write")
    public void writeObject(@RequestParam String bucketName, @RequestParam String objectName, @RequestBody String content) {
        storageService.writeObject(bucketName, objectName, content);
    }

    @DeleteMapping("/delete")
    public void deleteObject(@RequestParam String bucketName, @RequestParam String objectName) {
        storageService.deleteObject(bucketName, objectName);
    }
}
