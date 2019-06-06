// automatically generated by the FlatBuffers compiler, do not modify

namespace teamboss_msg
{

using System;
using FlatBuffers;

/// [Notify]伤害\损失HP通知
public sealed class BossData : Table {
  public static BossData GetRootAsBossData(ByteBuffer _bb) { return GetRootAsBossData(_bb, new BossData()); }
  public static BossData GetRootAsBossData(ByteBuffer _bb, BossData obj) { return (obj.__init(_bb.GetInt(_bb.Position) + _bb.Position, _bb)); }
  public BossData __init(int _i, ByteBuffer _bb) { bb_pos = _i; bb = _bb; return this; }

  public int BossReleaseSkillID { get { int o = __offset(4); return o != 0 ? bb.GetInt(o + bb_pos) : (int)0; } }
  public int BossComboCount { get { int o = __offset(6); return o != 0 ? bb.GetInt(o + bb_pos) : (int)0; } }
  public teamboss_msg.Vector BossStartAttackPos { get { return GetBossStartAttackPos(new teamboss_msg.Vector()); } }
  public teamboss_msg.Vector GetBossStartAttackPos(teamboss_msg.Vector obj) { int o = __offset(8); return o != 0 ? obj.__init(__indirect(o + bb_pos), bb) : null; }
  public teamboss_msg.Vector BossStartAttackDir { get { return GetBossStartAttackDir(new teamboss_msg.Vector()); } }
  public teamboss_msg.Vector GetBossStartAttackDir(teamboss_msg.Vector obj) { int o = __offset(10); return o != 0 ? obj.__init(__indirect(o + bb_pos), bb) : null; }
  public long BossStartAttackTimeStamp { get { int o = __offset(12); return o != 0 ? bb.GetLong(o + bb_pos) : (long)0; } }
  public teamboss_msg.BossSimpleData SimpleData { get { return GetSimpleData(new teamboss_msg.BossSimpleData()); } }
  public teamboss_msg.BossSimpleData GetSimpleData(teamboss_msg.BossSimpleData obj) { int o = __offset(14); return o != 0 ? obj.__init(__indirect(o + bb_pos), bb) : null; }

  public static Offset<BossData> CreateBossData(FlatBufferBuilder builder,
      int bossReleaseSkillID = 0,
      int bossComboCount = 0,
      Offset<teamboss_msg.Vector> bossStartAttackPosOffset = default(Offset<teamboss_msg.Vector>),
      Offset<teamboss_msg.Vector> bossStartAttackDirOffset = default(Offset<teamboss_msg.Vector>),
      long bossStartAttackTimeStamp = 0,
      Offset<teamboss_msg.BossSimpleData> simpleDataOffset = default(Offset<teamboss_msg.BossSimpleData>)) {
    builder.StartObject(6);
    BossData.AddBossStartAttackTimeStamp(builder, bossStartAttackTimeStamp);
    BossData.AddSimpleData(builder, simpleDataOffset);
    BossData.AddBossStartAttackDir(builder, bossStartAttackDirOffset);
    BossData.AddBossStartAttackPos(builder, bossStartAttackPosOffset);
    BossData.AddBossComboCount(builder, bossComboCount);
    BossData.AddBossReleaseSkillID(builder, bossReleaseSkillID);
    return BossData.EndBossData(builder);
  }

  public static void StartBossData(FlatBufferBuilder builder) { builder.StartObject(6); }
  public static void AddBossReleaseSkillID(FlatBufferBuilder builder, int bossReleaseSkillID) { builder.AddInt(0, bossReleaseSkillID, 0); }
  public static void AddBossComboCount(FlatBufferBuilder builder, int bossComboCount) { builder.AddInt(1, bossComboCount, 0); }
  public static void AddBossStartAttackPos(FlatBufferBuilder builder, Offset<teamboss_msg.Vector> bossStartAttackPosOffset) { builder.AddOffset(2, bossStartAttackPosOffset.Value, 0); }
  public static void AddBossStartAttackDir(FlatBufferBuilder builder, Offset<teamboss_msg.Vector> bossStartAttackDirOffset) { builder.AddOffset(3, bossStartAttackDirOffset.Value, 0); }
  public static void AddBossStartAttackTimeStamp(FlatBufferBuilder builder, long bossStartAttackTimeStamp) { builder.AddLong(4, bossStartAttackTimeStamp, 0); }
  public static void AddSimpleData(FlatBufferBuilder builder, Offset<teamboss_msg.BossSimpleData> simpleDataOffset) { builder.AddOffset(5, simpleDataOffset.Value, 0); }
  public static Offset<BossData> EndBossData(FlatBufferBuilder builder) {
    int o = builder.EndObject();
    return new Offset<BossData>(o);
  }
  public static void FinishBossDataBuffer(FlatBufferBuilder builder, Offset<BossData> offset) { builder.Finish(offset.Value); }
};


}
