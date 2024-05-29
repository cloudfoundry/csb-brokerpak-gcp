package org.cloudfoundry.spring.storage.service;

import com.google.cloud.storage.Blob;
import com.google.cloud.storage.BlobId;
import com.google.cloud.storage.BlobInfo;
import com.google.cloud.storage.Storage;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.nio.charset.StandardCharsets;

@Service
public class StorageService {

    @Autowired
    private final Storage storage;

    @Autowired
    public StorageService(Storage storage) {
        this.storage = storage;
    }

    public String readObject(String bucketName, String objectName) {
        Blob blob = storage.get(BlobId.of(bucketName, objectName));
        return new String(blob.getContent(), StandardCharsets.UTF_8);
    }

    public void writeObject(String bucketName, String objectName, String content) {
        BlobId blobId = BlobId.of(bucketName, objectName);
        BlobInfo blobInfo = BlobInfo.newBuilder(blobId).build();
        storage.create(blobInfo, content.getBytes(StandardCharsets.UTF_8));
    }

    public void deleteObject(String bucketName, String objectName) {
        storage.delete(BlobId.of(bucketName, objectName));
    }
}
