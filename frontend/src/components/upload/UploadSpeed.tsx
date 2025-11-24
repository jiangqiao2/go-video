import React from 'react';
import { Typography } from 'antd';

const { Text } = Typography;

interface UploadSpeedProps {
    speed: number; // bytes per second
}

export const UploadSpeed: React.FC<UploadSpeedProps> = ({ speed }) => {
    const formatSpeed = (bytesPerSecond: number) => {
        if (bytesPerSecond === 0) return '0 KB/s';

        const k = 1024;
        const sizes = ['B/s', 'KB/s', 'MB/s', 'GB/s'];
        const i = Math.floor(Math.log(bytesPerSecond) / Math.log(k));

        return `${parseFloat((bytesPerSecond / Math.pow(k, i)).toFixed(2))} ${sizes[i]}`;
    };

    return (
        <Text type="secondary" style={{ fontSize: '12px' }}>
            {formatSpeed(speed)}
        </Text>
    );
};

export default UploadSpeed;
