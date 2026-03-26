package com.generated.temporal.model;

public enum FileStorageType {
    SHARED_FS("shared_fs"),
    DATABASE("database"),
    S3("s3");

    private final String value;

    FileStorageType(String value) { this.value = value; }

    public String value() { return value; }

    public static FileStorageType fromValue(String value) {
        for (FileStorageType item : values()) {
            if (item.value.equalsIgnoreCase(value)) return item;
        }
        throw new IllegalArgumentException("Unsupported value: " + value);
    }
}
