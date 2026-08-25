"""
AI 智能聊天服务 - DeepSeek 大模型
支持多轮对话、多会话管理、内容管理和发布
"""

import os
import re
import sys
import json
import uuid
import time
import threading
import logging

# 配置日志
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s [%(levelname)s] %(message)s',
    stream=sys.stdout
)
logger = logging.getLogger(__name__)
from flask import Flask, request, jsonify
from flask_cors import CORS
from dotenv import load_dotenv
import requests
from pymongo import MongoClient, ASCENDING, DESCENDING
from pymongo.errors import ConnectionFailure, ServerSelectionTimeoutError

# 加载环境变量
load_dotenv()

app = Flask(__name__)
CORS(app)

# ==================== 配置 ====================

# DeepSeek API 配置
API_KEY = os.getenv("DEEPSEEK_API_KEY", "")
BASE_URL = "https://api.deepseek.com/v1"

# 服务配置
HOST = os.getenv("FLASK_HOST", "0.0.0.0")
PORT = int(os.getenv("FLASK_PORT", 5000))
DEBUG = os.getenv("FLASK_DEBUG", "true").lower() == "true"

# MongoDB 配置
MONGO_URI = os.getenv("MONGO_URI", "mongodb://localhost:27017")
MONGO_DB = os.getenv("MONGO_DB", "aiagent")

# 主服务后端地址（用于发布动态/知识等实际操作）
BACKEND_URL = os.getenv("BACKEND_URL", "http://localhost:8080")

# 模型配置
MODELS = {
    "chat-assistant": "deepseek-chat",
    "system-assistant": "deepseek-chat",
}

DEFAULT_MODEL = "chat-assistant"

# ==================== 提示词配置 ====================

PROMPTS_DIR = os.path.join(os.path.dirname(__file__), "prompts")
os.makedirs(PROMPTS_DIR, exist_ok=True)

def _load_prompt(filename):
    """从 prompts 文件夹加载提示词，如果文件不存在则返回空字符串"""
    filepath = os.path.join(PROMPTS_DIR, filename)
    if os.path.exists(filepath):
        with open(filepath, 'r', encoding='utf-8') as f:
            return f.read().strip()
    return ""

# 聊天助手基础系统提示词
CHAT_ASSISTANT_PROMPT = _load_prompt("chat_assistant.txt")
if not CHAT_ASSISTANT_PROMPT:
    CHAT_ASSISTANT_PROMPT = """你是一个专业的农业技术 AI 助手，专注于帮助用户解答农业相关问题。
你可以：
- 解答种植、养殖技术问题
- 提供病虫害防治建议
- 分享农业知识和管理经验
- 撰写农业相关文章和动态
请用热情专业的态度回答用户问题。"""

# 系统助手基础系统提示词
SYSTEM_ASSISTANT_PROMPT = _load_prompt("system_assistant.txt")
if not SYSTEM_ASSISTANT_PROMPT:
    SYSTEM_ASSISTANT_PROMPT = """你是一个系统级 AI 助手，具有以下能力：

1. **文件操作**：你可以请求读取项目中的文件（使用 read_file: 路径），以及生成代码写入文件（使用 write_file: 路径 + 内容）。
2. **内容发布**：你可以根据用户的主题，生成并发布动态(post)或知识库文章(knowledge)到平台。
3. **代码生成**：你可以为项目生成完整的代码文件。
4. **项目管理**：你可以查看项目结构，分析依赖关系。

当用户需要你操作文件时，请使用以下格式：
- 读取文件：`[ACTION:read_file]文件路径[/ACTION]`
- 写入文件：`[ACTION:write_file]文件路径|文件内容[/ACTION]`
- 列出目录：`[ACTION:list_dir]目录路径[/ACTION]`

项目工作区路径为：%WORKSPACE_ROOT%
""" % os.path.expandvars("${USERPROFILE}") if os.name == 'nt' else os.path.expandvars("${HOME}")

# 默认系统提示词文件内容（当文件不存在时自动创建）
DEFAULT_PROMPTS = {
    "chat_assistant.txt": """你是一个专业的农业技术 AI 助手，专注于帮助用户解答农业相关问题。
你可以：
- 解答种植、养殖技术问题
- 提供病虫害防治建议
- 分享农业知识和管理经验
- 撰写农业相关文章和动态
请用热情专业的态度回答用户问题。""",

    "system_assistant.txt": """你是一个系统级 AI 助手，可以执行以下操作：

## 1. 文件操作
你可以使用特定格式请求操作项目文件：
- 读取文件：在回复中使用格式 `[ACTION:read_file]文件路径[/ACTION]` 来读取文件内容
- 生成代码：使用代码块清晰展示要生成的代码，并标注文件路径
- 列出目录：使用 `[ACTION:list_dir]目录路径[/ACTION]` 来查看目录结构

## 2. 动态与知识库操作
你可以帮助用户在平台上生成内容：
- 生成动态(post)：撰写引人入胜的社交动态
- 生成知识(knowledge)：撰写专业的知识库文章
- 生成公告(notice)：撰写正式通知/公告

## 3. 技术能力
- 分析项目代码结构
- 生成完整的代码实现
- 解答技术问题
- 代码审查和建议

请在回复时：
- 如果是代码生成，确保代码完整可运行
- 如果是文件操作，清晰地说明操作意图
- 如果涉及内容发布，确保内容质量"""
}

# 自动创建默认提示词文件
for filename, content in DEFAULT_PROMPTS.items():
    filepath = os.path.join(PROMPTS_DIR, filename)
    if not os.path.exists(filepath):
        with open(filepath, 'w', encoding='utf-8') as f:
            f.write(content)

# ==================== 内容存储配置 ====================

CONTENT_DIR = os.path.join(os.path.dirname(__file__), "contents")
os.makedirs(CONTENT_DIR, exist_ok=True)

# ==================== 会话管理（MongoDB） ====================

class SessionManager:
    """会话管理器 - MongoDB 持久化 + 用户隔离"""

    def __init__(self, mongo_uri, db_name, expire_hours=48, max_history=30):
        self.expire_hours = expire_hours
        self.max_history = max_history
        self.lock = threading.Lock()

        try:
            self.client = MongoClient(mongo_uri, serverSelectionTimeoutMS=5000)
            self.client.admin.command('ping')
            self.db = self.client[db_name]
            self.collection = self.db["sessions"]

            # 创建索引：user_id + updated_at 联合查询，TTL 自动过期
            self.collection.create_index([("user_id", ASCENDING), ("updated_at", DESCENDING)])
            self.collection.create_index("updated_at", expireAfterSeconds=expire_hours * 3600)
            print(f"[MongoDB] 已连接: {mongo_uri}, 数据库: {db_name}")
        except (ConnectionFailure, ServerSelectionTimeoutError) as e:
            print(f"[MongoDB] 连接失败: {e}，使用文件存储作为后备")
            self.client = None
            self._fallback = {}
            self._fallback_file = os.path.join(CONTENT_DIR, "sessions_fallback.json")
            self._load_fallback()

    def _load_fallback(self):
        try:
            if os.path.exists(self._fallback_file):
                with open(self._fallback_file, 'r', encoding='utf-8') as f:
                    self._fallback = json.load(f)
        except:
            self._fallback = {}

    def _save_fallback(self):
        try:
            with open(self._fallback_file, 'w', encoding='utf-8') as f:
                json.dump(self._fallback, f, ensure_ascii=False, indent=2)
        except Exception as e:
            print(f"[SessionManager] 保存后备存储失败: {e}")

    def _to_doc(self, session):
        """将会话对象转为 MongoDB 文档"""
        return {
            "_id": session["id"],
            "user_id": session.get("user_id", "anonymous"),
            "title": session["title"],
            "messages": session["messages"],
            "model": session.get("model", DEFAULT_MODEL),
            "created_at": session["created_at"],
            "updated_at": session["updated_at"]
        }

    def _from_doc(self, doc):
        """将 MongoDB 文档转为会话对象"""
        if doc is None:
            return None
        return {
            "id": doc["_id"],
            "user_id": doc.get("user_id", "anonymous"),
            "title": doc.get("title", ""),
            "messages": doc.get("messages", []),
            "model": doc.get("model", DEFAULT_MODEL),
            "created_at": doc.get("created_at", time.time()),
            "updated_at": doc.get("updated_at", time.time())
        }

    def create_session(self, user_id="anonymous", title=None):
        session_id = str(uuid.uuid4())
        now = time.time()
        session = {
            "id": session_id,
            "user_id": user_id,
            "title": title or f"新对话",
            "messages": [],
            "model": DEFAULT_MODEL,
            "created_at": now,
            "updated_at": now
        }
        if self.client:
            try:
                self.collection.insert_one(self._to_doc(session))
            except Exception as e:
                print(f"[SessionManager] MongoDB 写入失败: {e}")
        else:
            with self.lock:
                self._fallback[session_id] = session
                self._save_fallback()
        return session

    def get_session(self, session_id, user_id=None):
        session = None
        if self.client:
            try:
                doc = self.collection.find_one({"_id": session_id})
                session = self._from_doc(doc)
            except Exception as e:
                print(f"[SessionManager] MongoDB 读取失败: {e}")
        else:
            with self.lock:
                session = self._fallback.get(session_id)

        if not session:
            return None

        # 验证用户所有权
        if user_id and session.get("user_id") != user_id:
            return None

        # 检查过期
        if time.time() - session["updated_at"] > self.expire_hours * 3600:
            self.delete_session(session_id)
            return None

        return session

    def update_session(self, session_id, messages=None, model=None, title=None, user_id=None):
        session = self.get_session(session_id, user_id)
        if not session:
            return False

        now = time.time()
        update_fields = {"updated_at": now}
        if messages is not None:
            update_fields["messages"] = messages[-self.max_history:]
        if model is not None:
            update_fields["model"] = model
        if title is not None:
            update_fields["title"] = title

        if self.client:
            try:
                self.collection.update_one({"_id": session_id}, {"$set": update_fields})
            except Exception as e:
                print(f"[SessionManager] MongoDB 更新失败: {e}")
        else:
            with self.lock:
                if session_id in self._fallback:
                    self._fallback[session_id].update(update_fields)
                    self._save_fallback()
        return True

    def delete_session(self, session_id, user_id=None):
        session = self.get_session(session_id, user_id)
        if not session:
            return False

        if self.client:
            try:
                self.collection.delete_one({"_id": session_id})
            except Exception as e:
                print(f"[SessionManager] MongoDB 删除失败: {e}")
        else:
            with self.lock:
                if session_id in self._fallback:
                    del self._fallback[session_id]
                    self._save_fallback()
        return True

    def list_sessions(self, user_id=None):
        session_list = []
        if self.client:
            try:
                query = {}
                if user_id:
                    query["user_id"] = user_id
                docs = self.collection.find(query).sort("updated_at", DESCENDING)
                for doc in docs:
                    session_list.append({
                        "id": doc["_id"],
                        "user_id": doc.get("user_id", "anonymous"),
                        "title": doc.get("title", ""),
                        "message_count": len(doc.get("messages", [])),
                        "created_at": doc.get("created_at", 0),
                        "updated_at": doc.get("updated_at", 0),
                        "model": doc.get("model", DEFAULT_MODEL)
                    })
            except Exception as e:
                print(f"[SessionManager] MongoDB 列表查询失败: {e}")
        else:
            with self.lock:
                now = time.time()
                for sid, s in list(self._fallback.items()):
                    if user_id and s.get("user_id") != user_id:
                        continue
                    if now - s["updated_at"] > self.expire_hours * 3600:
                        del self._fallback[sid]
                        continue
                    session_list.append({
                        "id": sid,
                        "user_id": s.get("user_id", "anonymous"),
                        "title": s.get("title", ""),
                        "message_count": len(s.get("messages", [])),
                        "created_at": s["created_at"],
                        "updated_at": s["updated_at"],
                        "model": s.get("model", DEFAULT_MODEL)
                    })
                self._save_fallback()
            session_list.sort(key=lambda x: x["updated_at"], reverse=True)
        return session_list

    def add_message(self, session_id, role, content, user_id=None):
        session = self.get_session(session_id, user_id)
        if not session:
            return None

        now = time.time()
        message = {"role": role, "content": content}

        if self.client:
            try:
                self.collection.update_one(
                    {"_id": session_id},
                    {
                        "$push": {"messages": message},
                        "$set": {"updated_at": now}
                    }
                )
                # 首条 user 消息作标题
                if role == "user" and session.get("messages") is None or len(session.get("messages", [])) == 0:
                    new_title = content[:30] + ("..." if len(content) > 30 else "")
                    self.collection.update_one(
                        {"_id": session_id},
                        {"$set": {"title": new_title}}
                    )
                # 裁剪历史
                session_doc = self.collection.find_one({"_id": session_id})
                if session_doc and len(session_doc.get("messages", [])) > self.max_history:
                    trimmed = session_doc["messages"][-self.max_history:]
                    self.collection.update_one(
                        {"_id": session_id},
                        {"$set": {"messages": trimmed}}
                    )
            except Exception as e:
                print(f"[SessionManager] MongoDB 添加消息失败: {e}")
        else:
            with self.lock:
                if session_id in self._fallback:
                    self._fallback[session_id]["messages"].append(message)
                    self._fallback[session_id]["updated_at"] = now
                    if role == "user" and len(self._fallback[session_id]["messages"]) == 1:
                        self._fallback[session_id]["title"] = content[:30] + ("..." if len(content) > 30 else "")
                    if len(self._fallback[session_id]["messages"]) > self.max_history:
                        self._fallback[session_id]["messages"] = self._fallback[session_id]["messages"][-self.max_history:]
                    self._save_fallback()
        return message

    def clear_history(self, session_id, user_id=None):
        session = self.get_session(session_id, user_id)
        if not session:
            return False

        if self.client:
            try:
                self.collection.update_one(
                    {"_id": session_id},
                    {"$set": {"messages": [], "updated_at": time.time()}}
                )
            except Exception as e:
                print(f"[SessionManager] MongoDB 清空失败: {e}")
        else:
            with self.lock:
                if session_id in self._fallback:
                    self._fallback[session_id]["messages"] = []
                    self._fallback[session_id]["updated_at"] = time.time()
                    self._save_fallback()
        return True

session_manager = SessionManager(MONGO_URI, MONGO_DB)

# ==================== 内容管理 ====================

class ContentManager:
    """内容管理器 - 保存和管理草稿"""

    def __init__(self):
        self.contents = {}
        self.lock = threading.Lock()
        self._load_from_disk()

    def _get_file_path(self):
        return os.path.join(CONTENT_DIR, "drafts.json")

    def _load_from_disk(self):
        """从磁盘加载内容"""
        filepath = self._get_file_path()
        if os.path.exists(filepath):
            try:
                with open(filepath, 'r', encoding='utf-8') as f:
                    data = json.load(f)
                    self.contents = {k: v for k, v in data.items()}
            except Exception as e:
                print(f"加载内容失败: {e}")
                self.contents = {}

    def _save_to_disk(self):
        """保存内容到磁盘"""
        filepath = self._get_file_path()
        try:
            with open(filepath, 'w', encoding='utf-8') as f:
                json.dump(self.contents, f, ensure_ascii=False, indent=2)
        except Exception as e:
            print(f"保存内容失败: {e}")

    def create_content(self, user_id, title, content, content_type, metadata=None):
        """创建新内容（草稿）"""
        with self.lock:
            content_id = str(uuid.uuid4())
            self.contents[content_id] = {
                "id": content_id,
                "user_id": user_id,
                "title": title,
                "content": content,
                "type": content_type,
                "status": "draft",  # draft, pending, published, rejected
                "metadata": metadata or {},
                "created_at": time.time(),
                "updated_at": time.time(),
                "published_at": None
            }
            self._save_to_disk()
            return self.contents[content_id]

    def get_content(self, content_id):
        return self.contents.get(content_id)

    def update_content(self, content_id, user_id, title=None, content=None, metadata=None):
        """更新内容"""
        with self.lock:
            if content_id in self.contents:
                c = self.contents[content_id]
                # 只能修改自己的草稿
                if c["user_id"] != user_id:
                    return None, "无权限修改"
                if c["status"] not in ["draft", "rejected"]:
                    return None, "只能修改草稿或被退回的内容"
                if title is not None:
                    c["title"] = title
                if content is not None:
                    c["content"] = content
                if metadata is not None:
                    c["metadata"] = metadata
                c["updated_at"] = time.time()
                self._save_to_disk()
                return c, None
            return None, "内容不存在"

    def submit_for_review(self, content_id, user_id):
        """提交审核"""
        with self.lock:
            if content_id in self.contents:
                c = self.contents[content_id]
                if c["user_id"] != user_id:
                    return None, "无权限"
                if c["status"] not in ["draft", "rejected"]:
                    return None, "只能提交草稿或被退回的内容"
                c["status"] = "pending"
                c["updated_at"] = time.time()
                self._save_to_disk()
                return c, None
            return None, "内容不存在"

    def publish_content(self, content_id):
        """发布内容（管理员操作）"""
        with self.lock:
            if content_id in self.contents:
                c = self.contents[content_id]
                c["status"] = "published"
                c["published_at"] = time.time()
                c["updated_at"] = time.time()
                self._save_to_disk()
                return c, None
            return None, "内容不存在"

    def reject_content(self, content_id, reason=None):
        """退回内容（管理员操作）"""
        with self.lock:
            if content_id in self.contents:
                c = self.contents[content_id]
                c["status"] = "rejected"
                c["reject_reason"] = reason
                c["updated_at"] = time.time()
                self._save_to_disk()
                return c, None
            return None, "内容不存在"

    def delete_content(self, content_id, user_id):
        """删除内容"""
        with self.lock:
            if content_id in self.contents:
                c = self.contents[content_id]
                if c["user_id"] != user_id and c["user_id"] != "admin":
                    return False, "无权限"
                del self.contents[content_id]
                self._save_to_disk()
                return True, None
            return False, "内容不存在"

    def list_contents(self, user_id=None, status=None, content_type=None):
        """列出内容"""
        with self.lock:
            result = []
            for c in self.contents.values():
                if user_id and c["user_id"] != user_id:
                    continue
                if status and c["status"] != status:
                    continue
                if content_type and c["type"] != content_type:
                    continue
                result.append(c)
            result.sort(key=lambda x: x["updated_at"], reverse=True)
            return result

content_manager = ContentManager()

# ==================== 权限定义 ====================

# 用户角色权限
ROLE_PERMISSIONS = {
    "user": {
        "can_create_content": True,
        "can_publish_directly": False,  # 需要审核
        "content_types": ["knowledge", "post"],
        "can_view_all": False
    },
    "expert": {
        "can_create_content": True,
        "can_publish_directly": True,  # 专家可直接发布
        "content_types": ["knowledge", "post"],
        "can_view_all": False
    },
    "admin": {
        "can_create_content": True,
        "can_publish_directly": True,
        "can_review": True,
        "content_types": ["knowledge", "post", "notice"],
        "can_view_all": True
    },
    "sysadmin": {
        "can_create_content": True,
        "can_publish_directly": True,
        "can_review": True,
        "can_delete_any": True,
        "content_types": ["knowledge", "post", "notice", "announcement"],
        "can_view_all": True
    }
}

def check_permission(role, permission):
    """检查用户权限"""
    if not role:
        return False
    perms = ROLE_PERMISSIONS.get(role, {})
    return perms.get(permission, False)

def can_publish_without_review(role, content_type):
    """检查是否可以免审核发布"""
    if not role:
        return False
    perms = ROLE_PERMISSIONS.get(role, {})
    if not perms.get("can_publish_directly", False):
        return False
    return content_type in perms.get("content_types", [])

# ==================== DeepSeek API 调用 ====================

def call_deepseek(messages, model=DEFAULT_MODEL, temperature=0.7, max_tokens=4096):
    if not API_KEY:
        return {"error": "API_KEY 未设置，请检查 .env 配置文件中的 DEEPSEEK_API_KEY"}

    url = f"{BASE_URL}/chat/completions"
    headers = {
        "Authorization": f"Bearer {API_KEY}",
        "Content-Type": "application/json"
    }
    payload = {
        "model": MODELS.get(model, model),
        "messages": messages,
        "temperature": temperature,
        "max_tokens": max_tokens
    }

    try:
        response = requests.post(url, headers=headers, json=payload, timeout=120)
        response.raise_for_status()
        return response.json()
    except requests.exceptions.Timeout:
        return {"error": "请求超时，请稍后重试"}
    except requests.exceptions.RequestException as e:
        error_msg = str(e)
        try:
            error_data = e.response.json() if e.response else {}
            if "error" in error_data:
                error_msg = error_data["error"].get("message", error_msg)
        except:
            pass
        return {"error": f"请求失败: {error_msg}"}


# ==================== API 接口 ====================

@app.route("/api/health", methods=["GET"])
def health_check():
    return jsonify({
        "status": "ok",
        "service": "AI 智能聊天 (DeepSeek)",
        "version": "4.0.0",
        "config": {
            "api_key_status": "已配置" if API_KEY else "未配置",
            "available_models": ["chat-assistant", "system-assistant"]
        }
    })


@app.route("/api/health/deepseek", methods=["GET"])
def health_check_deepseek():
    """独立的外部 API 连接检测（耗时操作，按需调用）"""
    if not API_KEY:
        return jsonify({"status": "skipped", "message": "API Key 未配置"})

    try:
        test_response = requests.post(
            f"{BASE_URL}/chat/completions",
            headers={
                "Authorization": f"Bearer {API_KEY}",
                "Content-Type": "application/json"
            },
            json={
                "model": DEFAULT_MODEL,
                "messages": [{"role": "user", "content": "Hi"}],
                "max_tokens": 5
            },
            timeout=10
        )
        if test_response.status_code == 200:
            return jsonify({"status": "ok", "message": "DeepSeek API 连接正常"})
        else:
            return jsonify({"status": "error", "message": f"状态码: {test_response.status_code}"}), 502
    except requests.exceptions.Timeout:
        return jsonify({"status": "error", "message": "连接超时"}), 504
    except requests.exceptions.ConnectionError:
        return jsonify({"status": "error", "message": "无法连接到 DeepSeek API，请检查网络或代理设置"}), 502
    except Exception as e:
        return jsonify({"status": "error", "message": str(e)}), 500


# ==================== 权限接口 ====================

@app.route("/api/permissions", methods=["GET"])
def get_permissions():
    """获取当前用户的权限"""
    role = request.args.get("role", "user")
    return jsonify({
        "role": role,
        "permissions": ROLE_PERMISSIONS.get(role, {}),
        "content_types": ROLE_PERMISSIONS.get(role, {}).get("content_types", [])
    })


# ==================== 会话管理接口 ====================

def _normalize_user_id(user_id_str):
    """将URL参数中的user_id转为合适的类型（与MongoDB存储的int保持一致）"""
    if not user_id_str:
        return None
    try:
        return int(user_id_str)  # 转为整数，与创建时存储的类型一致
    except (ValueError, TypeError):
        return user_id_str  # 非数字保持原样（如"anonymous"）

@app.route("/api/sessions", methods=["GET"])
def get_sessions():
    user_id = _normalize_user_id(request.args.get("user_id", ""))
    sessions = session_manager.list_sessions(user_id)
    return jsonify({"sessions": sessions})


@app.route("/api/sessions", methods=["POST"])
def create_session():
    data = request.get_json() or {}
    title = data.get("title")
    user_id = data.get("user_id", "anonymous")
    session = session_manager.create_session(user_id, title)
    return jsonify({
        "id": session["id"],
        "title": session["title"],
        "model": session["model"]
    })


@app.route("/api/sessions/<session_id>", methods=["GET"])
def get_session_detail(session_id):
    user_id = _normalize_user_id(request.args.get("user_id", ""))
    session = session_manager.get_session(session_id, user_id)
    if not session:
        return jsonify({"error": "会话不存在或已过期"}), 404
    return jsonify({
        "id": session["id"],
        "title": session["title"],
        "messages": session["messages"],
        "model": session["model"],
        "created_at": session["created_at"],
        "updated_at": session["updated_at"]
    })


@app.route("/api/sessions/<session_id>", methods=["PUT"])
def update_session_info(session_id):
    data = request.get_json() or {}
    user_id = data.get("user_id", "")
    session = session_manager.get_session(session_id, user_id if user_id else None)
    if not session:
        return jsonify({"error": "会话不存在或已过期"}), 404

    session_manager.update_session(
        session_id,
        title=data.get("title"),
        model=data.get("model"),
        user_id=user_id if user_id else None
    )
    return jsonify({"message": "更新成功"})


@app.route("/api/sessions/<session_id>", methods=["DELETE"])
def delete_session_api(session_id):
    user_id = _normalize_user_id(request.args.get("user_id", ""))
    if session_manager.delete_session(session_id, user_id):
        return jsonify({"message": "删除成功"})
    return jsonify({"error": "会话不存在"}), 404


@app.route("/api/sessions/<session_id>/clear", methods=["POST"])
def clear_session_history(session_id):
    data = request.get_json() or {}
    user_id = data.get("user_id", "")
    if session_manager.clear_history(session_id, user_id if user_id else None):
        return jsonify({"message": "历史已清空"})
    return jsonify({"error": "会话不存在"}), 404


# ==================== 内容管理接口 ====================

@app.route("/api/contents", methods=["GET"])
def get_contents():
    """获取内容列表"""
    user_id = _normalize_user_id(request.args.get("user_id"))
    status = request.args.get("status")
    content_type = request.args.get("type")

    # 如果指定了user_id，只返回该用户的内容
    contents = content_manager.list_contents(user_id, status, content_type)
    return jsonify({"contents": contents})


@app.route("/api/contents", methods=["POST"])
def create_content():
    """创建新内容"""
    data = request.get_json()

    if not data:
        return jsonify({"error": "缺少参数"}), 400

    user_id = data.get("user_id", "anonymous")
    role = data.get("role", "user")
    title = data.get("title", "")
    content = data.get("content", "")
    content_type = data.get("type", "knowledge")
    metadata = data.get("metadata", {})

    # 检查权限
    if not check_permission(role, "can_create_content"):
        return jsonify({"error": "您没有创建内容的权限"}), 403

    if content_type not in ROLE_PERMISSIONS.get(role, {}).get("content_types", []):
        return jsonify({"error": f"您没有权限创建该类型内容"}), 403

    # 创建内容
    result = content_manager.create_content(user_id, title, content, content_type, metadata)

    # 如果可以免审核发布，直接发布
    if can_publish_without_review(role, content_type):
        content_manager.publish_content(result["id"])

    return jsonify({
        "content": result,
        "auto_published": can_publish_without_review(role, content_type)
    })


@app.route("/api/contents/<content_id>", methods=["GET"])
def get_content_detail(content_id):
    """获取内容详情"""
    content = content_manager.get_content(content_id)
    if not content:
        return jsonify({"error": "内容不存在"}), 404
    return jsonify({"content": content})


@app.route("/api/contents/<content_id>", methods=["PUT"])
def update_content_api(content_id):
    """更新内容"""
    data = request.get_json()
    if not data:
        return jsonify({"error": "缺少参数"}), 400

    user_id = data.get("user_id", "anonymous")
    role = data.get("role", "user")
    title = data.get("title")
    content = data.get("content")
    metadata = data.get("metadata")

    # 管理员可以修改任何内容
    if not check_permission(role, "can_view_all"):
        content = content_manager.get_content(content_id)
        if content and content["user_id"] != user_id:
            return jsonify({"error": "无权限修改"}), 403

    result, error = content_manager.update_content(content_id, user_id, title, content, metadata)
    if error:
        return jsonify({"error": error}), 400
    return jsonify({"content": result})


@app.route("/api/contents/<content_id>", methods=["DELETE"])
def delete_content_api(content_id):
    """删除内容"""
    data = request.get_json() or {}
    user_id = data.get("user_id", "anonymous")
    role = data.get("role", "user")

    # 检查权限
    content = content_manager.get_content(content_id)
    if not content:
        return jsonify({"error": "内容不存在"}), 404

    if not check_permission(role, "can_delete_any") and content["user_id"] != user_id:
        return jsonify({"error": "无权限删除"}), 403

    success, error = content_manager.delete_content(content_id, user_id)
    if not success:
        return jsonify({"error": error}), 400
    return jsonify({"message": "删除成功"})


@app.route("/api/contents/<content_id>/submit", methods=["POST"])
def submit_content_for_review(content_id):
    """提交内容审核"""
    data = request.get_json() or {}
    user_id = data.get("user_id", "anonymous")

    result, error = content_manager.submit_for_review(content_id, user_id)
    if error:
        return jsonify({"error": error}), 400
    return jsonify({"content": result})


@app.route("/api/contents/<content_id>/publish", methods=["POST"])
def publish_content_api(content_id):
    """发布内容（管理员）"""
    data = request.get_json() or {}
    role = data.get("role", "user")

    if not check_permission(role, "can_review"):
        return jsonify({"error": "您没有发布权限"}), 403

    result, error = content_manager.publish_content(content_id)
    if error:
        return jsonify({"error": error}), 400
    return jsonify({"content": result})


@app.route("/api/contents/<content_id>/reject", methods=["POST"])
def reject_content_api(content_id):
    """退回内容（管理员）"""
    data = request.get_json() or {}
    role = data.get("role", "user")
    reason = data.get("reason", "")

    if not check_permission(role, "can_review"):
        return jsonify({"error": "您没有审核权限"}), 403

    result, error = content_manager.reject_content(content_id, reason)
    if error:
        return jsonify({"error": error}), 400
    return jsonify({"content": result})


# ==================== 聊天接口 ====================

@app.route("/api/chat", methods=["POST"])
def chat():
    data = request.get_json()

    if not data or "message" not in data:
        return jsonify({"error": "缺少 message 参数"}), 400

    message = data["message"]
    session_id = data.get("session_id")
    model = data.get("model", DEFAULT_MODEL)
    temperature = float(data.get("temperature", 0.7))
    user_id = data.get("user_id", "anonymous")
    token = data.get("token", "")

    logger.info(f"[CHAT] stream request: model={model}, token={'present' if token else 'MISSING'}, user_id={user_id}")

    if session_id:
        session = session_manager.get_session(session_id, user_id)
        if not session:
            return jsonify({"error": "会话不存在或已过期，请创建新会话"}), 404
    else:
        session = session_manager.create_session(user_id)

    session_id = session["id"]

    session_manager.add_message(session_id, "user", message, user_id)
    session = session_manager.get_session(session_id, user_id)
    messages = session["messages"] if session else [{"role": "user", "content": message}]

    # 构建带系统提示词的消息列表
    chat_messages = _build_messages_with_prompt(model, messages)

    result = call_deepseek(chat_messages, MODELS.get(model, "deepseek-chat"), temperature)

    if "error" in result:
        return jsonify(result), 500

    try:
        choices = result.get("choices", [])
        if choices:
            reply = choices[0]["message"]["content"]

            # Parse and execute actions for system-assistant
            actions_results = []
            stored_reply = reply
            if model == "system-assistant":
                clean_reply, actions_results = _parse_and_execute_actions(reply, user_id, token)
                stored_reply = clean_reply

            session_manager.add_message(session_id, "assistant", stored_reply, user_id)

            return jsonify({
                "session_id": session_id,
                "reply": stored_reply,
                "model": model,
                "actions_executed": actions_results,
                "usage": result.get("usage", {})
            })
    except (KeyError, IndexError):
        return jsonify({"error": "响应格式解析失败", "raw": result}), 500

    return jsonify({"error": "未知错误", "raw": result}), 500


def _build_messages_with_prompt(model, messages):
    """根据模型类型，在消息列表前注入对应的系统提示词。
    每次都使用 files 读取的最新版本，替换旧的 system 消息。"""
    system_prompt = ""
    if model == "chat-assistant":
        system_prompt = CHAT_ASSISTANT_PROMPT
    elif model == "system-assistant":
        system_prompt = SYSTEM_ASSISTANT_PROMPT

    if system_prompt:
        msgs = list(messages)
        # 移除旧的 system 消息，确保每次都使用最新提示词
        msgs = [m for m in msgs if m.get("role") != "system"]
        return [{"role": "system", "content": system_prompt}] + msgs
    return list(messages)


def _call_backend_api(endpoint, payload, token):
    """Call the main backend API (port 8080) to create/publish content."""
    logger.info(f"[ACTION] Calling backend: POST {BACKEND_URL}{endpoint} | payload={payload} | has_token={'yes' if token else 'NO'}")
    headers = {
        "Content-Type": "application/json",
        "Authorization": token
    }
    try:
        resp = requests.post(
            f"{BACKEND_URL}{endpoint}",
            json=payload,
            headers=headers,
            timeout=15
        )
        logger.info(f"[ACTION] Backend response: HTTP {resp.status_code} | body={resp.text[:200]}")
        if resp.status_code == 200:
            return resp.json(), None
        else:
            err_msg = f"HTTP {resp.status_code}"
            try:
                err_body = resp.json()
                err_msg = err_body.get("error") or err_body.get("msg") or err_msg
            except Exception:
                pass
            return None, err_msg
    except requests.exceptions.ConnectionError:
        logger.error(f"[ACTION] Connection refused: {BACKEND_URL} — 主服务未启动？")
        return None, f"无法连接到主服务 ({BACKEND_URL})，请确认服务是否启动"
    except Exception as e:
        logger.error(f"[ACTION] Request failed: {e}")
        return None, str(e)


def _parse_and_execute_actions(reply, user_id, token=""):
    """Parse [ACTION:...] tags in AI response and execute them via main backend API.
    Returns (cleaned_reply, actions_results_list)."""
    # 兼容 AI 可能的格式差异：空格、换行等
    action_pattern = re.compile(r'\[ACTION\s*:\s*(\w+)\]\s*(.*?)\s*\[/ACTION\]', re.DOTALL)
    actions_results = []

    # 先检查是否有 ACTION 标签
    has_actions = bool(action_pattern.search(reply))
    logger.info(f"[ACTION] Parsing reply ({len(reply)} chars), has ACTION tag: {has_actions}, token present: {'yes' if token else 'NO'}")

    if not has_actions:
        # 打印回复末尾，帮助诊断 AI 是否生成了正确格式
        tail = reply[-500:] if len(reply) > 500 else reply
        logger.info(f"[ACTION] No ACTION tag found. Reply tail (last 500 chars):\n---\n{tail}\n---")

    def replacer(match):
        action_type = match.group(1)
        action_content = match.group(2).strip()
        logger.info(f"[ACTION] Matched action_type={action_type}, content_len={len(action_content)}")

        if action_type == 'publish_post':
            parts = action_content.split('|', 1)
            title = parts[0].strip() if len(parts) > 0 else ""
            body = parts[1].strip() if len(parts) > 1 else ""

            # 如果只有一个部分（没有竖线分隔），尝试用第一行作为标题
            if not body and title:
                lines = title.split('\n', 1)
                title = lines[0].strip()
                body = lines[1].strip() if len(lines) > 1 else ""
                logger.info(f"[ACTION] publish_post: no '|' separator, using first line as title")

            if title and body:
                logger.info(f"[ACTION] publish_post: title='{title[:50]}...', body_len={len(body)}")
                payload = {"title": title, "content": body}
                result, error = _call_backend_api("/api/post/", payload, token)
                if error:
                    actions_results.append({
                        "action": "publish_post",
                        "title": title,
                        "status": "failed",
                        "error": error
                    })
                    return f"\n\n---\n> **Status**: Failed to publish post -- **{title}**\n> Error: {error}"
                else:
                    actions_results.append({
                        "action": "publish_post",
                        "title": title,
                        "content_id": result.get("data", {}).get("id", "unknown"),
                        "status": "published"
                    })
                    return f"\n\n---\n> **Status**: Post published successfully -- **{title}**"
            else:
                logger.warning(f"[ACTION] publish_post: missing title or body. parts={len(parts)}, title='{title[:80] if title else ''}'")
                return match.group(0)

        elif action_type == 'publish_knowledge':
            parts = action_content.split('|', 1)
            title = parts[0].strip() if len(parts) > 0 else ""
            body = parts[1].strip() if len(parts) > 1 else ""

            if not body and title:
                lines = title.split('\n', 1)
                title = lines[0].strip()
                body = lines[1].strip() if len(lines) > 1 else ""
                logger.info(f"[ACTION] publish_knowledge: no '|' separator, using first line as title")

            if title and body:
                logger.info(f"[ACTION] publish_knowledge: title='{title[:50]}...', body_len={len(body)}")
                payload = {"title": title, "content": body}
                result, error = _call_backend_api("/api/knowledge/", payload, token)
                if error:
                    actions_results.append({
                        "action": "publish_knowledge",
                        "title": title,
                        "status": "failed",
                        "error": error
                    })
                    return f"\n\n---\n> **Status**: Failed to publish knowledge -- **{title}**\n> Error: {error}"
                else:
                    actions_results.append({
                        "action": "publish_knowledge",
                        "title": title,
                        "content_id": result.get("data", {}).get("id", "unknown"),
                        "status": "published"
                    })
                    return f"\n\n---\n> **Status**: Knowledge article published successfully -- **{title}**"
            else:
                logger.warning(f"[ACTION] publish_knowledge: missing title or body. parts={len(parts)}")
                return match.group(0)

        # Unknown action, leave as-is
        logger.warning(f"[ACTION] Unknown action type: {action_type}")
        return match.group(0)

    clean_reply = action_pattern.sub(replacer, reply)
    logger.info(f"[ACTION] Parsing done. actions_executed={len(actions_results)}")
    return clean_reply, actions_results


@app.route("/api/chat/stream", methods=["POST"])
def chat_stream():
    data = request.get_json()

    if not data or "message" not in data:
        return jsonify({"error": "缺少 message 参数"}), 400

    message = data["message"]
    session_id = data.get("session_id")
    model = data.get("model", DEFAULT_MODEL)
    temperature = float(data.get("temperature", 0.7))
    user_id = data.get("user_id", "anonymous")
    token = data.get("token", "")

    logger.info(f"[CHAT] stream request: model={model}, token={'present' if token else 'MISSING'}, user_id={user_id}")

    if session_id:
        session = session_manager.get_session(session_id, user_id)
        if not session:
            return jsonify({"error": "会话不存在或已过期"}), 404
    else:
        session = session_manager.create_session(user_id)

    session_id = session["id"]

    session_manager.add_message(session_id, "user", message, user_id)
    session = session_manager.get_session(session_id, user_id)
    messages = session["messages"] if session else [{"role": "user", "content": message}]

    # 构建带系统提示词的消息列表
    chat_messages = _build_messages_with_prompt(model, messages)

    if not API_KEY:
        return jsonify({"error": "API_KEY 未配置"}), 500

    url = f"{BASE_URL}/chat/completions"
    headers = {
        "Authorization": f"Bearer {API_KEY}",
        "Content-Type": "application/json"
    }
    payload = {
        "model": MODELS.get(model, "deepseek-chat"),
        "messages": chat_messages,
        "temperature": temperature,
        "stream": True
    }

    def generate():
        full_reply = ""
        try:
            response = requests.post(url, headers=headers, json=payload, timeout=120, stream=True)
            response.raise_for_status()

            for line in response.iter_lines():
                if line:
                    line_text = line.decode('utf-8')
                    if line_text.startswith("data: "):
                        data_str = line_text[6:]
                        if data_str == "[DONE]":
                            break
                        try:
                            data_json = json.loads(data_str)
                            delta = data_json.get("choices", [{}])[0].get("delta", {})
                            content = delta.get("content", "")
                            if content:
                                full_reply += content
                                yield f"data: {json.dumps({'content': content}, ensure_ascii=False)}\n\n"
                        except json.JSONDecodeError:
                            continue

            # Parse and execute actions in the AI response
            actions_results = []
            stored_reply = full_reply
            if full_reply and model == "system-assistant":
                clean_reply, actions_results = _parse_and_execute_actions(full_reply, user_id, token)
                stored_reply = clean_reply

            if stored_reply:
                session_manager.add_message(session_id, "assistant", stored_reply, user_id)

            # Yield action results before done
            for ar in actions_results:
                yield f"data: {json.dumps({'action': ar}, ensure_ascii=False)}\n\n"

            yield f"data: {json.dumps({'done': True, 'session_id': session_id, 'actions_executed': len(actions_results)}, ensure_ascii=False)}\n\n"

        except Exception as e:
            yield f"data: {json.dumps({'error': str(e)}, ensure_ascii=False)}\n\n"

    from flask import Response
    return Response(
        generate(),
        mimetype='text/event-stream',
        headers={
            'Cache-Control': 'no-cache',
            'Connection': 'keep-alive',
            'X-Accel-Buffering': 'no'
        }
    )


# ==================== AI 发布动态内容接口 ====================

# AI 生成内容的系统提示词模板
CONTENT_GENERATE_PROMPTS = {
    "post": """你是一个专业的内容创作者。请根据用户的主题和要求，生成一篇高质量的动态/帖子内容。

要求：
1. 内容要有吸引力，适合社交平台发布
2. 可以适当使用 emoji 和活泼的语气（但不要过度）
3. 结构清晰，适当分段
4. 如果是技术类内容，保持专业准确
5. 长度适中，300-1000字为宜
6. 只输出正文内容，不要包含"标题："等前缀""",

    "knowledge": """你是一个专业知识库作者。请根据用户的主题和要求，生成一篇知识库文章。

要求：
1. 内容严谨专业，结构清晰
2. 包含必要的背景说明、核心知识点、实践建议
3. 使用标题层级组织内容（用 # ## ### 表示）
4. 适当使用示例和代码块（如适用）
5. 长度适中，500-2000字
6. 直接输出正文内容""",

    "notice": """你是一个官方公告撰写者。请根据用户的主题和要求，生成一篇正式通知/公告。

要求：
1. 语言正式得体，符合官方公告风格
2. 包含通知对象、事由、具体要求、截止时间等要素（如适用）
3. 结构清晰，逻辑严谨
4. 长度适中，200-800字
5. 直接输出正文内容"""
}


@app.route("/api/chat/publish", methods=["POST"])
def chat_publish():
    """AI 根据话题生成内容并直接发布为动态"""
    data = request.get_json()

    if not data or "topic" not in data:
        return jsonify({"error": "缺少 topic 参数，请提供发布主题"}), 400

    topic = data["topic"]
    content_type = data.get("type", "post")  # post, knowledge, notice
    user_id = data.get("user_id", "anonymous")
    role = data.get("role", "user")
    instructions = data.get("instructions", "")
    model = data.get("model", DEFAULT_MODEL)

    # 检查创建内容权限
    if not check_permission(role, "can_create_content"):
        return jsonify({"error": "您没有创建内容的权限"}), 403

    if content_type not in ROLE_PERMISSIONS.get(role, {}).get("content_types", []):
        return jsonify({"error": f"您的角色 {role} 没有权限创建 {content_type} 类型内容"}), 403

    if not API_KEY:
        return jsonify({"error": "API_KEY 未配置，无法调用 AI 生成内容"}), 500

    # 构建系统提示词
    system_prompt = CONTENT_GENERATE_PROMPTS.get(content_type, CONTENT_GENERATE_PROMPTS["post"])

    # 构建用户消息
    user_message = f"主题：{topic}"
    if instructions:
        user_message += f"\n额外要求：{instructions}"

    messages = [
        {"role": "system", "content": system_prompt},
        {"role": "user", "content": user_message}
    ]

    # 1. AI 生成标题
    title_messages = [
        {"role": "system", "content": "你是一个标题撰写专家。请根据内容主题生成一个简洁有力的标题，15字以内。直接输出标题，不要任何前缀。"},
        {"role": "user", "content": f"为以下主题生成标题：{topic}"}
    ]
    title_result = call_deepseek(title_messages, model, temperature=0.5, max_tokens=50)
    if "error" in title_result:
        # 标题生成失败，使用主题作为标题
        title = topic[:30]
    else:
        try:
            title = title_result["choices"][0]["message"]["content"].strip().strip('"').strip("'")
            if len(title) > 50:
                title = title[:50]
        except (KeyError, IndexError):
            title = topic[:30]

    # 2. AI 生成正文内容
    content_result = call_deepseek(messages, model, temperature=0.8, max_tokens=4096)
    if "error" in content_result:
        return jsonify(content_result), 500

    try:
        generated_content = content_result["choices"][0]["message"]["content"]
    except (KeyError, IndexError):
        return jsonify({"error": "AI 响应格式解析失败"}), 500

    # 3. 创建内容（走 ContentManager）
    created = content_manager.create_content(
        user_id=user_id,
        title=title,
        content=generated_content,
        content_type=content_type,
        metadata={
            "source": "ai_generated",
            "model": model,
            "topic": topic,
            "instructions": instructions,
            "generated_at": time.time()
        }
    )

    # 4. 按角色权限决定发布策略
    auto_published = False
    if can_publish_without_review(role, content_type):
        result, _ = content_manager.publish_content(created["id"])
        if result:
            auto_published = True
    else:
        # 普通用户提交审核
        content_manager.submit_for_review(created["id"], user_id)

    return jsonify({
        "content": {
            "id": created["id"],
            "title": title,
            "content": generated_content,
            "type": content_type,
            "status": created["status"],
            "auto_published": auto_published,
            "usage": content_result.get("usage", {})
        },
        "message": "内容已发布" if auto_published else "内容已保存并提交审核"
    })


@app.route("/api/chat/generate", methods=["POST"])
def chat_generate_content():
    """AI 生成内容但不发布，返回生成结果供用户预览修改"""
    data = request.get_json()

    if not data or "topic" not in data:
        return jsonify({"error": "缺少 topic 参数"}), 400

    topic = data["topic"]
    content_type = data.get("type", "post")
    instructions = data.get("instructions", "")
    model = data.get("model", DEFAULT_MODEL)

    if not API_KEY:
        return jsonify({"error": "API_KEY 未配置"}), 500

    system_prompt = CONTENT_GENERATE_PROMPTS.get(content_type, CONTENT_GENERATE_PROMPTS["post"])
    user_message = f"主题：{topic}"
    if instructions:
        user_message += f"\n额外要求：{instructions}"

    messages = [
        {"role": "system", "content": system_prompt},
        {"role": "user", "content": user_message}
    ]

    result = call_deepseek(messages, model, temperature=0.8, max_tokens=4096)
    if "error" in result:
        return jsonify(result), 500

    try:
        generated_content = result["choices"][0]["message"]["content"]
    except (KeyError, IndexError):
        return jsonify({"error": "AI 响应格式解析失败"}), 500

    return jsonify({
        "content": generated_content,
        "topic": topic,
        "type": content_type,
        "usage": result.get("usage", {})
    })


# ==================== 模型列表接口 ====================

@app.route("/api/models", methods=["GET"])
def list_models():
    return jsonify({
        "models": [
            {"id": "chat-assistant", "name": "聊天助手", "description": "农业技术智能问答，解答种植养殖等专业问题"},
            {"id": "system-assistant", "name": "系统助手", "description": "系统级AI助手，支持文件操作、代码生成和内容自动发布"}
        ],
        "default": DEFAULT_MODEL
    })


# ==================== 系统助手工作区接口 ====================

# 安全的工作区根目录（系统助手只能在这些路径下操作）
ALLOWED_WORKSPACES = []

def _init_workspaces():
    """初始化允许访问的工作区目录"""
    global ALLOWED_WORKSPACES
    default_workspaces = []
    # 添加当前项目所在目录的父目录作为工作区
    project_root = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
    if os.path.isdir(project_root):
        default_workspaces.append(os.path.normpath(project_root))
    # 环境变量配置的额外工作区
    env_workspace = os.getenv("AI_WORKSPACE", "")
    if env_workspace:
        for p in env_workspace.split(";" if os.name == "nt" else ":"):
            p = os.path.normpath(p.strip())
            if p and os.path.isdir(p) and p not in default_workspaces:
                default_workspaces.append(p)
    ALLOWED_WORKSPACES = default_workspaces

_init_workspaces()


def _is_path_safe(filepath):
    """检查文件路径是否在允许的工作区内"""
    abs_path = os.path.normpath(os.path.abspath(filepath))
    for ws in ALLOWED_WORKSPACES:
        try:
            if abs_path.startswith(ws + os.sep) or abs_path == ws:
                return True
        except:
            pass
    return False


@app.route("/api/prompts", methods=["GET"])
def get_prompts():
    """获取所有可用的提示词"""
    prompt_type = request.args.get("type", "all")
    result = {}
    if prompt_type in ("all", "chat-assistant"):
        result["chat-assistant"] = CHAT_ASSISTANT_PROMPT
    if prompt_type in ("all", "system-assistant"):
        result["system-assistant"] = SYSTEM_ASSISTANT_PROMPT
    return jsonify({"prompts": result})


@app.route("/api/prompts/<prompt_type>", methods=["PUT"])
def update_prompt(prompt_type):
    """更新指定类型的提示词"""
    if prompt_type not in ("chat-assistant", "system-assistant"):
        return jsonify({"error": "无效的提示词类型"}), 400

    data = request.get_json()
    if not data or "content" not in data:
        return jsonify({"error": "缺少 content 参数"}), 400

    filename = f"{prompt_type}.txt"
    filepath = os.path.join(PROMPTS_DIR, filename)
    with open(filepath, 'w', encoding='utf-8') as f:
        f.write(data["content"])

    # 重新加载到内存
    global CHAT_ASSISTANT_PROMPT, SYSTEM_ASSISTANT_PROMPT
    if prompt_type == "chat-assistant":
        CHAT_ASSISTANT_PROMPT = data["content"]
    else:
        SYSTEM_ASSISTANT_PROMPT = data["content"]

    return jsonify({"message": "提示词已更新"})


@app.route("/api/workspace/read", methods=["POST"])
def workspace_read():
    """系统助手：读取工作区文件"""
    data = request.get_json()
    if not data or "path" not in data:
        return jsonify({"error": "缺少 path 参数"}), 400

    filepath = data["path"]
    # 支持相对路径（相对于工作区根目录）
    if not os.path.isabs(filepath):
        # 使用第一个工作区作为默认
        if not ALLOWED_WORKSPACES:
            return jsonify({"error": "未配置工作区"}), 500
        filepath = os.path.join(ALLOWED_WORKSPACES[0], filepath)

    if not _is_path_safe(filepath):
        return jsonify({"error": "不允许访问该路径"}), 403

    if not os.path.exists(filepath):
        return jsonify({"error": "文件不存在"}), 404

    if os.path.isdir(filepath):
        # 列出目录内容
        try:
            items = []
            for item in sorted(os.listdir(filepath)):
                item_path = os.path.join(filepath, item)
                items.append({
                    "name": item,
                    "type": "dir" if os.path.isdir(item_path) else "file",
                    "size": os.path.getsize(item_path) if os.path.isfile(item_path) else 0
                })
            return jsonify({"type": "directory", "path": filepath, "items": items})
        except Exception as e:
            return jsonify({"error": str(e)}), 500

    # 读取文件内容
    try:
        # 限制文件大小
        file_size = os.path.getsize(filepath)
        if file_size > 5 * 1024 * 1024:  # 5MB
            return jsonify({"error": "文件过大，无法读取"}), 413

        with open(filepath, 'r', encoding='utf-8') as f:
            content = f.read()

        return jsonify({
            "type": "file",
            "path": filepath,
            "name": os.path.basename(filepath),
            "size": file_size,
            "content": content
        })
    except UnicodeDecodeError:
        return jsonify({"error": "无法以文本方式读取此文件（可能是二进制文件）"}), 400
    except Exception as e:
        return jsonify({"error": str(e)}), 500


@app.route("/api/workspace/write", methods=["POST"])
def workspace_write():
    """系统助手：写入工作区文件"""
    data = request.get_json()
    if not data or "path" not in data or "content" not in data:
        return jsonify({"error": "缺少 path 或 content 参数"}), 400

    filepath = data["path"]
    if not os.path.isabs(filepath):
        if not ALLOWED_WORKSPACES:
            return jsonify({"error": "未配置工作区"}), 500
        filepath = os.path.join(ALLOWED_WORKSPACES[0], filepath)

    if not _is_path_safe(filepath):
        return jsonify({"error": "不允许写入该路径"}), 403

    try:
        # 确保目录存在
        os.makedirs(os.path.dirname(filepath), exist_ok=True)

        with open(filepath, 'w', encoding='utf-8') as f:
            f.write(data["content"])

        return jsonify({
            "message": "文件写入成功",
            "path": filepath,
            "size": len(data["content"])
        })
    except Exception as e:
        return jsonify({"error": str(e)}), 500


@app.route("/api/workspace/list", methods=["GET"])
def workspace_list():
    """系统助手：获取工作区目录树"""
    if not ALLOWED_WORKSPACES:
        return jsonify({"workspaces": [], "message": "未配置工作区"})

    result = []
    for ws in ALLOWED_WORKSPACES:
        result.append({
            "path": ws,
            "name": os.path.basename(ws) if ws else "workspace",
            "exists": os.path.exists(ws)
        })
    return jsonify({"workspaces": result})


# ==================== 启动服务 ====================

if __name__ == "__main__":
    import sys

    # Windows 下 debug reloader 会导致服务无响应，自动关闭 reloader
    use_reloader = DEBUG and sys.platform != "win32" 
    """
        DEBUG and sys.platform != "win32"
    """


    print("=" * 60)
    print("   AI 智能聊天服务 (DeepSeek)")
    print("   版本 4.0.0 - 聊天助手 & 系统助手")
    print("=" * 60)
    print(f"  服务地址: http://{HOST}:{PORT}")
    print(f"  API 配置: {'已配置' if API_KEY else '未配置'}")
    print(f"  Debug 模式: {'开启' if DEBUG else '关闭'}")
    print(f"  热重载: {'开启' if use_reloader else '关闭 (Windows 自动禁用)'}")
    print(f"  提示词目录: {PROMPTS_DIR}")
    print(f"  工作区: {ALLOWED_WORKSPACES}")
    print("-" * 60)
    print("  助手类型:")
    print("  - 聊天助手 (chat-assistant): 农业技术问答")
    print("  - 系统助手 (system-assistant): 文件操作 + 代码生成 + 内容发布")
    print("-" * 60)
    print("  接口说明:")
    print("  - GET  /api/health              - 健康检查")
    print("  - GET  /api/health/deepseek     - DeepSeek API 连通性检测")
    print("  - GET  /api/permissions         - 获取权限")
    print("  - GET  /api/models              - 获取可选助手")
    print("  - GET  /api/prompts             - 获取提示词")
    print("  - GET  /api/workspace/list      - 查看工作区")
    print("  - POST /api/workspace/read      - 读取文件")
    print("  - POST /api/workspace/write     - 写入文件")
    print("  - POST /api/chat                - 发送消息(非流式)")
    print("  - POST /api/chat/stream         - 流式聊天(SSE)")
    print("  - POST /api/chat/publish        - AI生成并发布内容")
    print("  - POST /api/chat/generate       - AI仅生成内容(预览)")
    print("=" * 60)
    print()

    app.run(
        host=HOST,
        port=PORT,
        debug=DEBUG,
        use_reloader=use_reloader
    )
