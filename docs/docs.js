// Mengri Flow 文档查看器
// 基于文件树导航的文档阅读器

// 全局变量
let documents = [];  // 所有文档列表
let currentDocument = null;  // 当前选中的文档

// 文件名到文档类型的映射
const fileTypeMap = {
    // 需求文档
    'prd': 'requirement',
    '需求': 'requirement',
    'product': 'requirement',
    
    // 架构设计  
    'architecture': 'architecture',
    'arch': 'architecture',
    'design': 'architecture',
    '架构': 'architecture',
    '设计': 'architecture',
    'executor': 'architecture',
    '执行器': 'architecture',
    
    // 开发指南
    'guide': 'guide',
    '开发': 'guide',
    'development': 'guide',
    'plugin': 'guide',
    '插件': 'guide',
    'developer': 'guide',
    
    // 流程与规划
    'flow': 'plan',
    '流程图': 'plan',
    'journey': 'plan',
    '用户体验': 'plan',
    '页面': 'plan',
    '信息架构': 'architecture',
    'information': 'architecture'
};

// 类型配置
const typeConfig = {
    'requirement': {
        name: '需求文档',
        icon: 'fas fa-clipboard-list',
        badgeClass: 'type-requirement'
    },
    'architecture': {
        name: '架构设计',
        icon: 'fas fa-sitemap',
        badgeClass: 'type-architecture'
    },
    'guide': {
        name: '开发指南',
        icon: 'fas fa-book',
        badgeClass: 'type-guide'
    },
    'plan': {
        name: '流程规划',
        icon: 'fas fa-project-diagram',
        badgeClass: 'type-plan'
    },
    'default': {
        name: '文档',
        icon: 'fas fa-file-alt',
        badgeClass: 'type-default'
    }
};

// Markdown 转 HTML 的转换器
const converter = new showdown.Converter({
    tables: true,
    smoothLivePreview: true,
    strikethrough: true,
    tasklists: true,
    // 确保标题生成ID，便于锚点链接
    headerLevelStart: 1,
    parseImgDimensions: true,
    simplifiedAutoLink: true,
    literalMidWordUnderscores: true,
    // 自动为标题生成ID
    customizeHeaderId: true,
    // 更好的标题锚点生成
    ghCompatibleHeaderId: true,
    // 支持扩展（如果需要）
    extensions: []
});

// 从文件名推断文档类型
function inferTypeFromFilename(filename) {
    const lowerName = filename.toLowerCase();
    
    for (const [keyword, type] of Object.entries(fileTypeMap)) {
        if (lowerName.includes(keyword.toLowerCase())) {
            return type;
        }
    }
    
    return 'default';
}

// 生成友好的显示名称
function generateDisplayName(filename) {
    // 移除扩展名
    const nameWithoutExt = filename.replace(/\.md$/i, '');
    
    // 转换驼峰命名和下划线为空格
    let displayName = nameWithoutExt
        .replace(/([A-Z])/g, ' $1')
        .replace(/[-_]/g, ' ')
        .replace(/\s+/g, ' ')
        .trim();
    
    // 首字母大写
    displayName = displayName.split(' ')
        .map(word => word.charAt(0).toUpperCase() + word.slice(1))
        .join(' ');
    
    return displayName;
}

// 根据类型获取文件图标
function getFileIcon(filename) {
    const type = inferTypeFromFilename(filename);
    
    const iconMap = {
        'requirement': 'fas fa-file-alt',
        'architecture': 'fas fa-sitemap',
        'guide': 'fas fa-book',
        'plan': 'fas fa-project-diagram',
        'default': 'fas fa-file'
    };
    
    return iconMap[type] || iconMap.default;
}

// 扫描 docs 目录获取文件列表
async function scanDocsDirectory() {
    try {
        console.log('开始扫描 docs 目录...');
        
        // 已知的文件列表（作为备选方案）
        const knownFiles = [
            'PRD.md',
            'architecture-design.md',
            'executor-design-v2.md',
            'developer-guide.md',
            'plugin-development-guide.md',
            'information-architecture.md',
            'user-journey-map.md',
            'page-flow.md'
        ];
        
        // 尝试动态获取文件列表
        let fileList = knownFiles;
        
        try {
            // 尝试获取目录列表文件（如果存在）
            const response = await fetch('./.filelist');
            if (response.ok) {
                const text = await response.text();
                const lines = text.split('\n').filter(line => line.trim().endsWith('.md'));
                if (lines.length > 0) {
                    fileList = lines.map(line => line.trim());
                }
            }
        } catch (dirError) {
            console.log('无法获取动态文件列表，使用预定义列表');
        }
        
        // 验证文件是否存在
        const validFiles = [];
        
        for (const filename of fileList) {
            try {
                const response = await fetch(filename);
                if (response.ok) {
                    const type = inferTypeFromFilename(filename);
                    const config = typeConfig[type] || typeConfig.default;
                    
                    // 获取文件信息
                    const fileSize = response.headers.get('content-length') || 'N/A';
                    const lastModified = response.headers.get('last-modified') || 
                                        new Date().toISOString().split('T')[0];
                    
                    validFiles.push({
                        id: filename.toLowerCase().replace(/[^a-z0-9]/g, '-'),
                        name: generateDisplayName(filename),
                        filename: filename,
                        displayName: generateDisplayName(filename),
                        filepath: filename,
                        type: type,
                        icon: config.icon,
                        badgeClass: config.badgeClass,
                        typeName: config.name,
                        lastModified: formatDate(lastModified),
                        size: formatFileSize(fileSize)
                    });
                    
                    console.log(`已添加文档: ${filename}`);
                }
            } catch (error) {
                console.log(`文档 ${filename} 无法访问:`, error.message);
            }
        }
        
        // 如果没有找到任何文件，使用默认列表（用于调试）
        if (validFiles.length === 0) {
            console.warn('未找到任何Markdown文档，显示示例文档');
            
            // 创建一个更小的、保证有效的文件列表
            const fallbackFiles = [
                '测试文档1.md',
                '架构设计文档.md',
                '开发指南.md'
            ];
            
            fallbackFiles.forEach(filename => {
                const type = inferTypeFromFilename(filename);
                const config = typeConfig[type] || typeConfig.default;
                
                validFiles.push({
                    id: filename.toLowerCase().replace(/[^a-z0-9]/g, '-'),
                    name: generateDisplayName(filename),
                    filename: filename,
                    displayName: generateDisplayName(filename),
                    filepath: '/test/' + filename,
                    type: type,
                    icon: config.icon,
                    badgeClass: config.badgeClass,
                    typeName: config.name,
                    lastModified: new Date().toISOString().split('T')[0],
                    size: '1KB'
                });
            });
            
            console.log('已创建', validFiles.length, '个示例文档');
        }
        
        // 按名称排序
        validFiles.sort((a, b) => a.name.localeCompare(b.name));
        
        documents = validFiles;
        console.log(`扫描完成，共找到 ${documents.length} 个文档`);
        return documents;
        
    } catch (error) {
        console.error('扫描文档目录时发生错误:', error);
        return [];
    }
}

// 格式化日期
function formatDate(dateString) {
    try {
        const date = new Date(dateString);
        return date.toISOString().split('T')[0];
    } catch {
        return new Date().toISOString().split('T')[0];
    }
}

// 格式化文件大小
function formatFileSize(bytes) {
    if (bytes === 'N/A' || isNaN(bytes)) return 'N/A';
    
    const b = parseInt(bytes);
    if (b < 1024) return b + ' B';
    if (b < 1024 * 1024) return (b / 1024).toFixed(1) + ' KB';
    return (b / (1024 * 1024)).toFixed(1) + ' MB';
}

// 初始化函数
async function init() {
    console.log('初始化文档查看器...');
    
    try {
        // 扫描文档目录
        await scanDocsDirectory();
        console.log('文档扫描完成，找到', documents.length, '个文档');
        
        // 渲染文件树
        renderFileTree();
        console.log('文件树渲染完成');
        
        // 设置事件监听器
        setupEventListeners();
        console.log('事件监听器设置完成');
        
        // 如果有文档，默认加载第一个
        if (documents.length > 0) {
            console.log('正在加载第一个文档:', documents[0].filename);
            await loadDocument(documents[0]);
        }
        
        console.log('文档查看器初始化完成');
    } catch (error) {
        console.error('初始化失败:', error);
        showErrorState(error.message);
    }
}

// 渲染文件树
function renderFileTree() {
    const fileTree = document.getElementById('file-tree');
    const fileCount = document.getElementById('file-count');
    
    if (!fileTree) return;
    
    // 更新文件数量
    if (fileCount) {
        fileCount.textContent = documents.length;
    }
    
    fileTree.innerHTML = '';
    
    if (documents.length === 0) {
        fileTree.innerHTML = `
            <div class="empty-state">
                <i class="fas fa-exclamation-circle"></i>
                <p>没有找到文档</p>
            </div>
        `;
        return;
    }
    
    documents.forEach(doc => {
        const fileItem = document.createElement('div');
        fileItem.className = 'file-item';
        fileItem.dataset.file = doc.filename;
        fileItem.innerHTML = `
            <i class="${doc.icon}"></i>
            <span class="file-name">${doc.displayName}</span>
            <span class="file-type-badge ${doc.badgeClass}">${doc.typeName}</span>
        `;
        
        fileItem.addEventListener('click', () => loadDocument(doc));
        fileTree.appendChild(fileItem);
    });
}

// 加载并显示文档
async function loadDocument(doc) {
    console.log('加载文档:', doc.filename);
    
    // 更新选中状态
    updateSelectedFile(doc.filename);
    
    // 更新标题
    document.getElementById('doc-title').innerHTML = `
        <i class="${doc.icon}"></i> ${doc.displayName}
    `;
    
    // 更新文件路径
    document.getElementById('file-path').textContent = doc.filename;
    
    // 显示加载状态
    document.getElementById('document-content').innerHTML = `
        <div class="empty-state">
            <i class="fas fa-spinner fa-spin"></i>
            <h3>正在加载文档...</h3>
            <p>请稍候</p>
        </div>
    `;
    
    try {
        // 使用 fetch API 获取 Markdown 文件内容
        const response = await fetch(doc.filename);
        if (!response.ok) {
            throw new Error(`HTTP错误! 状态码: ${response.status}`);
        }
        
        const markdownContent = await response.text();
        
        // 转换 Markdown 为 HTML
        const htmlContent = converter.makeHtml(markdownContent);
        
        // 美化代码块
        const formattedHtml = htmlContent
            .replace(/<pre><code class="language-(\w+)">/g, '<pre class="language-$1"><code class="language-$1">')
            .replace(/<code>/g, '<code class="language-plaintext">');
        
        // 渲染内容
        document.getElementById('document-content').innerHTML = `
            <div class="markdown-content">
                ${formattedHtml}
            </div>
        `;
        
        // 保存当前文档
        currentDocument = doc;
        
        // 滚动到顶部
        document.querySelector('.document-content').scrollTop = 0;
        
        // 处理章节目录链接（锚点）
        setupAnchorLinks();
        
        // 增强代码块功能（复制按钮等）
        enhanceCodeBlocks();
        
        // 调试日志
        console.log('章节目录链接已设置，找到', 
                   document.querySelectorAll('.markdown-content a[href^="#"]').length, 
                   '个锚点链接');
        console.log('代码块增强完成');
        
    } catch (error) {
        console.error('加载文档失败:', error);
        document.getElementById('document-content').innerHTML = `
            <div class="empty-state" style="color: #dc2626;">
                <i class="fas fa-exclamation-triangle"></i>
                <h3>加载文档失败</h3>
                <p>无法加载文档: ${doc.filename}</p>
                <p style="font-size: 0.85rem;">错误信息: ${error.message}</p>
                <button onclick="window.location.reload()" style="
                    background: #2563eb;
                    color: white;
                    border: none;
                    padding: 8px 16px;
                    border-radius: 4px;
                    cursor: pointer;
                    margin-top: 15px;
                    font-size: 0.9rem;
                ">
                    <i class="fas fa-sync-alt"></i> 重新加载
                </button>
            </div>
        `;
    }
}

// 更新选中的文件状态
function updateSelectedFile(filename) {
    // 移除所有 active 类
    document.querySelectorAll('.file-item').forEach(item => {
        item.classList.remove('active');
    });
    
    // 添加当前 active 类
    const currentItem = document.querySelector(`.file-item[data-file="${filename}"]`);
    if (currentItem) {
        currentItem.classList.add('active');
    }
}

// 搜索文档
function searchDocuments(query) {
    query = query.toLowerCase().trim();
    
    const fileTree = document.getElementById('file-tree');
    if (!fileTree) return;
    
    fileTree.innerHTML = '';
    
    if (!query) {
        // 清空搜索，显示所有文档
        documents.forEach(doc => {
            const fileItem = document.createElement('div');
            fileItem.className = 'file-item';
            fileItem.dataset.file = doc.filename;
            fileItem.innerHTML = `
                <i class="${doc.icon}"></i>
                <span class="file-name">${doc.displayName}</span>
                <span class="file-type-badge ${doc.badgeClass}">${doc.typeName}</span>
            `;
            
            fileItem.addEventListener('click', () => loadDocument(doc));
            fileTree.appendChild(fileItem);
        });
        return;
    }
    
    // 搜索过滤
    const searchResults = documents.filter(doc => {
        return doc.displayName.toLowerCase().includes(query) ||
               doc.type.toLowerCase().includes(query) ||
               doc.filename.toLowerCase().includes(query);
    });
    
    if (searchResults.length === 0) {
        fileTree.innerHTML = `
            <div class="empty-state">
                <i class="fas fa-search"></i>
                <p>未找到匹配的文档</p>
                <button onclick="clearSearch()" style="
                    background: #f1f5f9;
                    color: #64748b;
                    border: 1px solid #e2e8f0;
                    padding: 6px 12px;
                    border-radius: 4px;
                    cursor: pointer;
                    margin-top: 10px;
                    font-size: 0.85rem;
                ">
                    显示全部
                </button>
            </div>
        `;
        return;
    }
    
    searchResults.forEach(doc => {
        const fileItem = document.createElement('div');
        fileItem.className = 'file-item';
        fileItem.dataset.file = doc.filename;
        fileItem.innerHTML = `
            <i class="${doc.icon}"></i>
            <span class="file-name">${doc.displayName}</span>
            <span class="file-type-badge ${doc.badgeClass}">${doc.typeName}</span>
        `;
        
        fileItem.addEventListener('click', () => loadDocument(doc));
        fileTree.appendChild(fileItem);
    });
}

// 清除搜索
function clearSearch() {
    const searchInput = document.getElementById('search-input');
    if (searchInput) {
        searchInput.value = '';
    }
    const fileTree = document.getElementById('file-tree');
    if (fileTree) {
        fileTree.innerHTML = '';
        documents.forEach(doc => {
            const fileItem = document.createElement('div');
            fileItem.className = 'file-item';
            fileItem.dataset.file = doc.filename;
            fileItem.innerHTML = `
                <i class="${doc.icon}"></i>
                <span class="file-name">${doc.displayName}</span>
                <span class="file-type-badge ${doc.badgeClass}">${doc.typeName}</span>
            `;
            
            fileItem.addEventListener('click', () => loadDocument(doc));
            fileTree.appendChild(fileItem);
        });
    }
}

// 显示错误状态
function showErrorState(errorMessage) {
    const fileTree = document.getElementById('file-tree');
    if (fileTree) {
        fileTree.innerHTML = `
            <div class="empty-state" style="color: #dc2626;">
                <i class="fas fa-exclamation-triangle"></i>
                <h3>加载失败</h3>
                <p>无法扫描文档目录</p>
                <p style="font-size: 0.85rem;">错误详情: ${errorMessage}</p>
                <button onclick="window.location.reload()" style="
                    background: #2563eb;
                    color: white;
                    border: none;
                    padding: 8px 16px;
                    border-radius: 4px;
                    cursor: pointer;
                    margin-top: 10px;
                    font-size: 0.85rem;
                ">
                    <i class="fas fa-sync-alt"></i> 重新加载
                </button>
            </div>
        `;
    }
    
    const contentArea = document.getElementById('document-content');
    if (contentArea) {
        contentArea.innerHTML = `
            <div class="empty-state" style="color: #dc2626;">
                <i class="fas fa-exclamation-triangle"></i>
                <h3>系统错误</h3>
                <p>文档查看器初始化失败</p>
                <button onclick="window.location.reload()" style="
                    background: #2563eb;
                    color: white;
                    border: none;
                    padding: 8px 16px;
                    border-radius: 4px;
                    cursor: pointer;
                    margin-top: 10px;
                    font-size: 0.85rem;
                ">
                    <i class="fas fa-sync-alt"></i> 重新加载页面
                </button>
            </div>
        `;
    }
}

// 处理章节目录链接（锚点）
function setupAnchorLinks() {
    try {
        const contentArea = document.querySelector('.document-content');
        if (!contentArea) {
            console.warn('未找到文档内容区域');
            return;
        }
        
        // 为所有内部锚点链接添加点击事件
        const anchorLinks = contentArea.querySelectorAll('a[href^="#"]');
        console.log(`找到 ${anchorLinks.length} 个章节目录链接`);
        
        anchorLinks.forEach((link) => {
            // 移除默认的点击行为
            link.addEventListener('click', function(e) {
                e.preventDefault();
                
                const href = this.getAttribute('href');
                if (!href || !href.startsWith('#')) return;
                
                const targetId = href.substring(1);
                if (!targetId) {
                    console.warn('空的锚点ID');
                    return;
                }
                
                console.log(`点击章节目录链接: ${href} (ID: ${targetId})`);
                
                // 查找目标元素（支持多种方式）
                let targetElement = document.getElementById(targetId);
                
                if (!targetElement) {
                    // 尝试查找有name属性的元素
                    targetElement = contentArea.querySelector(`[name="${targetId}"]`);
                }
                
                if (!targetElement) {
                    // 尝试查找包含特定类或属性的元素
                    targetElement = contentArea.querySelector(`h1[id*="${targetId}"], h2[id*="${targetId}"], h3[id*="${targetId}"]`);
                }
                
                if (!targetElement) {
                    console.warn(`未找到锚点目标: ${targetId}`);
                    
                    // 尝试寻找近似匹配
                    const normalizedId = targetId.toLowerCase().replace(/[\s-]+/g, '-');
                    const allHeaders = contentArea.querySelectorAll('h1, h2, h3, h4, h5, h6');
                    
                    for (const header of allHeaders) {
                        const headerText = header.textContent.toLowerCase().replace(/[\s-]+/g, '-');
                        if (headerText.includes(normalizedId) || normalizedId.includes(headerText)) {
                            targetElement = header;
                            break;
                        }
                    }
                }
                
                if (targetElement) {
                    console.log(`滚动到目标: ${targetElement.tagName} - ${targetElement.textContent}`);
                    
                    // 平滑滚动到目标位置
                    targetElement.scrollIntoView({
                        behavior: 'smooth',
                        block: 'start'
                    });
                    
                    // 添加临时高亮效果
                    const originalBackground = targetElement.style.backgroundColor;
                    targetElement.style.backgroundColor = '#f0f9ff';
                    targetElement.style.transition = 'background-color 0.3s ease';
                    
                    setTimeout(() => {
                        targetElement.style.backgroundColor = originalBackground;
                        setTimeout(() => {
                            targetElement.style.transition = '';
                        }, 300);
                    }, 1000);
                    
                    // 更新URL（但不重新加载页面）
                    history.pushState(null, '', `#${targetId}`);
                } else {
                    console.error(`无法找到目标元素: ${targetId}`);
                    
                    // 显示用户提示
                    const errorMsg = document.createElement('div');
                    errorMsg.className = 'anchor-error';
                    errorMsg.innerHTML = `<span style="color:#dc2626; background:#fee2e2; padding: 5px 10px; border-radius: 4px; margin-top: 10px; display: inline-block;">
                        <i class="fas fa-exclamation-triangle"></i> 无法跳转到章节 "${targetId}"
                    </span>`;
                    
                    const parent = this.parentElement;
                    if (parent) {
                        parent.appendChild(errorMsg);
                        setTimeout(() => {
                            if (errorMsg.parentElement) {
                                errorMsg.parentElement.removeChild(errorMsg);
                            }
                        }, 3000);
                    }
                }
            });
        });
        
        // 监听URL中的哈希变化（支持浏览器后退按钮）
        window.addEventListener('hashchange', function() {
            const hash = window.location.hash.substring(1);
            if (hash) {
                console.log('URL哈希变化:', hash);
                const targetElement = document.getElementById(hash) || 
                                     contentArea.querySelector(`[name="${hash}"]`);
                if (targetElement) {
                    setTimeout(() => {
                        targetElement.scrollIntoView({ behavior: 'smooth', block: 'start' });
                    }, 100);
                }
            }
        });
        
        // 初始检查URL中的哈希
        if (window.location.hash) {
            const hash = window.location.hash.substring(1);
            const targetElement = document.getElementById(hash) || 
                                 contentArea.querySelector(`[name="${hash}"]`);
            if (targetElement) {
                setTimeout(() => {
                    targetElement.scrollIntoView({ behavior: 'smooth', block: 'start' });
                }, 300);
            }
        }
        
    } catch (error) {
        console.error('设置章节目录链接时发生错误:', error);
    }
}

// 增强代码块功能（复制按钮、语言标识等）
function enhanceCodeBlocks() {
    try {
        const contentArea = document.querySelector('.document-content');
        if (!contentArea) return;
        
        // 找到所有的代码块
        const codeBlocks = contentArea.querySelectorAll('pre');
        console.log(`找到 ${codeBlocks.length} 个代码块`);
        
        codeBlocks.forEach((pre, index) => {
            // 添加唯一的ID用于复制功能
            const blockId = `code-block-${index}-${Date.now()}`;
            
            // 获取语言类型
            let language = 'plaintext';
            for (const className of pre.classList) {
                if (className.startsWith('language-')) {
                    language = className.replace('language-', '');
                    break;
                }
            }
            
            // 添加复制按钮
            const copyButton = document.createElement('button');
            copyButton.className = 'code-copy-button';
            copyButton.innerHTML = '<i class="far fa-copy"></i> 复制';
            copyButton.setAttribute('data-block-id', blockId);
            copyButton.setAttribute('data-language', language);
            
            // 复制功能
            copyButton.addEventListener('click', async function() {
                const codeElement = pre.querySelector('code');
                const codeText = codeElement ? codeElement.textContent : pre.textContent;
                
                try {
                    await navigator.clipboard.writeText(codeText);
                    
                    // 更新按钮状态
                    const originalHTML = this.innerHTML;
                    this.innerHTML = '<i class="fas fa-check"></i> 已复制';
                    this.classList.add('copied');
                    
                    setTimeout(() => {
                        this.innerHTML = originalHTML;
                        this.classList.remove('copied');
                    }, 2000);
                    
                    console.log(`代码已复制 (${language})`);
                } catch (err) {
                    console.error('Clipboard API复制失败:', err);
                    
                    // 备用方案：使用textarea和execCommand
                    try {
                        const textarea = document.createElement('textarea');
                        textarea.value = codeText;
                        textarea.style.position = 'fixed';
                        textarea.style.opacity = '0';
                        document.body.appendChild(textarea);
                        textarea.select();
                        
                        // 使用execCommand作为备选
                        const success = document.execCommand('copy');
                        document.body.removeChild(textarea);
                        
                        if (success) {
                            console.log('使用execCommand复制成功');
                        } else {
                            console.error('execCommand也失败了');
                        }
                    } catch (execError) {
                        console.error('execCommand也失败了:', execError);
                    }
                    
                    this.innerHTML = '<i class="fas fa-check"></i> 已复制';
                    this.classList.add('copied');
                    
                    setTimeout(() => {
                        this.innerHTML = originalHTML;
                        this.classList.remove('copied');
                    }, 2000);
                }
            });
            
            pre.appendChild(copyButton);
            
            // 添加语言标签
            if (language !== 'plaintext') {
                const titleSpan = document.createElement('span');
                titleSpan.className = 'code-title';
                titleSpan.textContent = language.toUpperCase();
                
                // 在pre元素前插入语言标签
                pre.parentNode.insertBefore(titleSpan, pre);
            }
            
            // 高亮常见语法（简单版本）
            if (pre.querySelector('code')) {
                applyBasicSyntaxHighlighting(pre, language);
            }
        });
        
    } catch (error) {
        console.error('增强代码块功能时出错:', error);
    }
}

// 应用基本语法高亮
function applyBasicSyntaxHighlighting(pre, language) {
    const codeElement = pre.querySelector('code');
    if (!codeElement) return;
    
    // 这里只是一个简单的示例，实际可以使用专门的语法高亮库
    // 如 highlight.js 或 Prism.js
    
    // 对于演示目的，我们只做一些简单的替换
    if (language === 'javascript' || language === 'js') {
        codeElement.innerHTML = codeElement.innerHTML
            .replace(/\b(function|return|const|let|var|if|else|for|while|switch|case|default|break|continue|class|new|this|import|export|from|as)\b/g, '<span class="keyword">$1</span>')
            .replace(/\b(console|document|window|JSON|Math|Date|Array|String|Number)\b/g, '<span class="class-name">$1</span>')
            .replace(/(".+?"|'.+?')/g, '<span class="string">$1</span>')
            .replace(/\/\/.+$/gm, '<span class="comment">$&</span>')
            .replace(/\/\*[\s\S]*?\*\//g, '<span class="comment">$&</span>')
            .replace(/\b(\d+)\b/g, '<span class="number">$1</span>');
    }
    
    if (language === 'python' || language === 'py') {
        codeElement.innerHTML = codeElement.innerHTML
            .replace(/\b(def|class|return|if|elif|else|for|while|import|from|as|with|try|except|finally|raise|assert|yield|async|await|lambda)\b/g, '<span class="keyword">$1</span>')
            .replace(/(".+?"|\'.+?\')/g, '<span class="string">$1</span>')
            .replace(/#.+$/gm, '<span class="comment">$&</span>')
            .replace(/\b(\d+)\b/g, '<span class="number">$1</span>');
    }
    
    // 更多语言支持可以根据需要添加
    // 注意：这是简化版本，实际项目建议使用专门的语法高亮库
}

// 设置事件监听器
function setupEventListeners() {
    // 搜索功能
    const searchInput = document.getElementById('search-input');
    if (searchInput) {
        let searchTimeout;
        
        searchInput.addEventListener('input', function() {
            clearTimeout(searchTimeout);
            searchTimeout = setTimeout(() => {
                searchDocuments(this.value);
            }, 300);
        });
        
        // Enter 键搜索
        searchInput.addEventListener('keypress', function(e) {
            if (e.key === 'Enter') {
                e.preventDefault();
                searchDocuments(this.value);
            }
        });
    }
    
    // 添加键盘快捷键
    document.addEventListener('keydown', function(e) {
        // Ctrl/Cmd + K 聚焦搜索框
        if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
            e.preventDefault();
            const searchInput = document.getElementById('search-input');
            if (searchInput) {
                searchInput.focus();
                searchInput.select();
            }
        }
        
        // Ctrl/Cmd + F 聚焦搜索框
        if ((e.ctrlKey || e.metaKey) && e.key === 'f') {
            e.preventDefault();
            const searchInput = document.getElementById('search-input');
            if (searchInput) {
                searchInput.focus();
                searchInput.select();
            }
        }
    });
}

// 页面加载完成后初始化
document.addEventListener('DOMContentLoaded', init);

// 暴露全局函数以便调试
window.MengriDocs = {
    getDocuments: () => documents,
    loadDocument,
    searchDocuments,
    clearSearch
};