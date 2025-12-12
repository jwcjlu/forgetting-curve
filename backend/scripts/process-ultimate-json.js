const fs = require('fs');
const path = require('path');

// 读取 ultimate.json 文件
const filePath = path.join(__dirname, '..', 'ultimate.json');
const backupPath = path.join(__dirname, '..', 'ultimate.json.backup');

console.log('开始处理 ultimate.json 文件...');
console.log('文件路径:', filePath);

// 创建备份
console.log('创建备份文件...');
fs.copyFileSync(filePath, backupPath);
console.log('备份文件已创建:', backupPath);

// 读取 JSON 文件
console.log('读取 JSON 文件...');
const fileContent = fs.readFileSync(filePath, 'utf8');
const data = JSON.parse(fileContent);

console.log(`总共 ${Object.keys(data).length} 个单词`);

// 处理数据
let processedCount = 0;
let cambridgeOnlyCount = 0;
let totalUrlsRemoved = 0;

for (const [word, urls] of Object.entries(data)) {
  if (!Array.isArray(urls) || urls.length === 0) {
    continue;
  }

  // 查找 Cambridge 字典的 URL
  const cambridgeUrls = urls.filter(url => 
    url && typeof url === 'string' && url.indexOf('http://dictionary.cambridge.org') === 0
  );

  // 如果有 Cambridge 字典的 URL，只保留这些
  if (cambridgeUrls.length > 0) {
    const originalCount = urls.length;
    data[word] = cambridgeUrls;
    processedCount++;
    cambridgeOnlyCount++;
    totalUrlsRemoved += (originalCount - cambridgeUrls.length);
    
    // 如果多个 Cambridge URL，只保留第一个
    if (cambridgeUrls.length > 1) {
      data[word] = [cambridgeUrls[0]];
      totalUrlsRemoved += (cambridgeUrls.length - 1);
    }
  }
}

console.log('\n处理完成！');
console.log(`- 处理的单词数: ${processedCount}`);
console.log(`- 只保留 Cambridge 的单词数: ${cambridgeOnlyCount}`);
console.log(`- 移除的 URL 总数: ${totalUrlsRemoved}`);

// 保存处理后的文件
console.log('\n保存处理后的文件...');
const outputContent = JSON.stringify(data, null, 2);
fs.writeFileSync(filePath, outputContent, 'utf8');
console.log('文件已保存:', filePath);
console.log('\n完成！原文件已备份到:', backupPath);



