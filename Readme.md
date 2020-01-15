# S3 Download and SFTP Upload

## Environment Variables

 Name | Required | Default | Description |
|------|----------|---------|-------------|
| `AUTHORIZED_CONFIG_ID` | n | `-` | S3 Trigger notification name. Providing this ensures no other notifications are filtered out. Otherwise, all notifications will be processed |
| `SFTP_HOST` | y | `-` | Destination SFTP Host name, e.g `sftp.domain.com`  |
| `SFTP_PORT` | y | `22` | Destination SFTP Port  |
| `SFTP_USERNAME` | y | `-` | Destination SFTP User Name   |
| `SFTP_PASSWORD` | y | `-` | Destination SFTP Password   |
| `UPLOAD_PATH` | n | `/import-inbox/` | Destination SFTP default path relative to chroot  |

## Development

### Serverless

serverless is required for this project

```bash
npm i -g serverless
```

### Dependencies

This project uses gomod, dependencies should be auto resolved

## Deployment

```bash
make deploy
```
